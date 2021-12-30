package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	ServerAddr string `split_words:"true" default:":8080" required:"true"`
	RedisAddr  string `split_words:"true" default:"localhost:6379" required:"true"`
	RedisDB    int    `split_words:"true" default:"1" required:"true"`
	KafkaAddr  string `split_words:"true" default:"localhost:9092" required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("bank", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(log.Writer(), log.Writer(), log.Writer()))

	lis, err := net.Listen("tcp", cfg.ServerAddr)
	if err != nil {
		log.Fatalf("failed listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(),
	}

	server := grpc.NewServer(opts...)
	reflection.Register(server)

	repo := &Repository{
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.RedisAddr,
			DB:   cfg.RedisDB,
		}),
	}

	interop, err := NewInterop([]string{cfg.KafkaAddr}, repo)
	if err != nil {
		log.Fatalf("failed to create interop: %v", err)
	}
	ctx := context.Background()

	srv := &Server{}
	grpc_health_v1.RegisterHealthServer(server, srv)

	errc := make(chan error, 1)

	go func() {
		log.Printf("starting gRPC on %s", lis.Addr())
		if err := server.Serve(lis); err != nil {
			errc <- fmt.Errorf("failed to serve: %v", err)
		}
	}()

	go func() {
		log.Println("starting interop")
		if err := interop.Start(ctx); err != nil {
			errc <- fmt.Errorf("failed to start interop: %v", err)
		}
	}()

	select {
	case err := <-errc:
		log.Fatalf("failed: %v", err)
	}
}
