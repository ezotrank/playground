package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	// producer
	topicBankAccounts     = "bank.accounts"
	topicBankTransactions = "bank.transactions"

	// consumer
	topicWalletUsers        = "wallet.users"
	topicWalletTransactions = "wallet.transactions"

	// WithRetry retry
	topicWalletUsersRetry        = "wallet.users__bank-interop__retry"
	topicWalletUsersDLQ          = "wallet.users__bank-interop__dlq"
	topicWalletTransactionsRetry = "wallet.transactions__bank-interop__retry"
	topicWalletTransactionsDLQ   = "wallet.transactions__bank-interop__dlq"
)

type Config struct {
	ServerAddr   string `split_words:"true" default:":8080" required:"true"`
	RedisAddr    string `split_words:"true" default:"localhost:6379" required:"true"`
	RedisDB      int    `split_words:"true" default:"1" required:"true"`
	KafkaAddr    string `split_words:"true" default:"localhost:9092" required:"true"`
	ExternalAddr string `split_words:"true" default:"localhost:9999" required:"true"`
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
	producer := NewProducer(cfg.KafkaAddr)
	external := NewExternal(cfg.ExternalAddr)
	handler := NewHandler(repo, producer, external)

	err = CreateTopic([]string{cfg.KafkaAddr}, topicWalletUsersRetry, topicWalletUsersDLQ)
	if err != nil {
		log.Panicln("failed create topic:", err)
	}

	interop, err := NewInterop([]string{cfg.KafkaAddr}, Flow{
		rules: map[string]Rule{
			topicWalletUsers: {
				Handler:  handler.WalletUsersHandler,
				Attempts: 1,
				DLQ:      topicWalletUsersRetry,
			},
			topicWalletUsersRetry: {
				Handler:  handler.WalletUsersHandler,
				Attempts: 3,
				DLQ:      topicWalletUsersDLQ,
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to create interop: %v", err)
	}

	srv := &Server{}
	grpc_health_v1.RegisterHealthServer(server, srv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		errc := make(chan error, 1)

		go func() {
			log.Printf("starting gRPC on %s", lis.Addr())
			errc <- server.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			server.GracefulStop()
		case err := <-errc:
			return err
		}

		return nil
	})
	g.Go(func() error {
		log.Println("starting interop")
		return interop.Start(ctx)
	})

	go func() {
		<-ctx.Done()
		<-time.Tick(time.Second * 30)
		log.Panicln("shutdown is exceed timeout")
	}()

	if err := g.Wait(); err != nil {
		log.Panicln(err)
	}
}
