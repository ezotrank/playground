package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/ezotrank/interop"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/external"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/handler"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/producer"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/repository"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/server"
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

	gsrv := grpc.NewServer(opts...)
	reflection.Register(gsrv)

	hdlr := handler.NewHandler(
		repository.NewRepository(cfg.RedisAddr, cfg.RedisDB),
		producer.NewProducer(
			producer.Topics{
				BankAccountsTopic: topicBankAccounts,
			},
			cfg.KafkaAddr,
		),
		external.NewExternal(cfg.ExternalAddr),
	)

	inrpt, err := interop.NewInterop([]string{cfg.KafkaAddr}, interop.Flow{
		Rules: map[string]interop.Rule{
			topicWalletUsers: {
				Handler:  hdlr.WalletUsersHandler,
				Attempts: 1,
				DLQ:      topicWalletUsersRetry,
			},
			topicWalletUsersRetry: {
				Handler:  hdlr.WalletUsersHandler,
				Attempts: 3,
				DLQ:      topicWalletUsersDLQ,
			},
		},
	}, "bank-interop")
	if err != nil {
		log.Fatalf("failed to create interop: %v", err)
	}

	srv := &server.Server{}
	grpc_health_v1.RegisterHealthServer(gsrv, srv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		errc := make(chan error)

		go func() {
			log.Printf("starting gRPC on %s", lis.Addr())
			errc <- gsrv.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			gsrv.GracefulStop()
			return nil
		case err := <-errc:
			return err
		}
	})
	g.Go(func() error {
		log.Println("starting interop")
		return inrpt.Start(ctx)
	})

	// If some routine can't be stopped, the whole program will be terminated.
	go func() {
		<-ctx.Done()
		<-time.Tick(time.Second * 30)
		log.Panicln("shutdown is exceed timeout")
	}()

	if err := g.Wait(); err != nil {
		log.Panicln(err)
	}
}
