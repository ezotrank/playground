package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	pbbank "github.com/ezotrank/playground/saga-choreography/bank/gen"
	pb "github.com/ezotrank/playground/saga-choreography/wallet/gen"
)

const (
	// producer
	topicWalletUsers = "wallet.users"

	// consumer
	topicBankAccounts = "bank.accounts"
)

const consumerGroup = "wallet-interop"

func NewInterop(brokers []string, repo *Repository) (*Interop, error) {
	if err := CreateTopic(brokers, topicWalletUsers); err != nil {
		return nil, fmt.Errorf("failed to create topic %s: %w", topicWalletUsers, err)
	}

	return &Interop{
		repo: repo,
		usersWriter: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topicWalletUsers,
		},
		bankAccountsReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topicBankAccounts,
			GroupID: consumerGroup,
		}),
	}, nil
}

func CreateTopic(brokers []string, name string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker: %v", err)
	}

	// Hack to not get a "Not Available: the cluster is in the middle" error in WriteMessages.
	// Create or check if topics exist.
	if err := conn.CreateTopics(
		kafka.TopicConfig{
			Topic:             name,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	); err != nil {
		return fmt.Errorf("failed to create topic: %v", err)
	}

	return nil
}

type Interop struct {
	repo               *Repository
	usersWriter        *kafka.Writer
	bankAccountsReader *kafka.Reader
}

func (i *Interop) Start(ctx context.Context) error {
	errc := make(chan error, 1)

	go func() {
		if err := i.bankConsumer(ctx); err != nil {
			errc <- err
		}
	}()

	select {
	case <-ctx.Done():
		return i.shutdown(ctx)
	case err := <-errc:
		return err
	}
}

func (i *Interop) shutdown(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return i.bankAccountsReader.Close()
	})
	g.Go(func() error {
		return i.usersWriter.Close()
	})

	return g.Wait()
}

func (i *Interop) bankConsumer(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			ctx := context.Background()
			msg, err := i.bankAccountsReader.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch message: %v", err)
			}

			var account pbbank.Account
			if err := proto.Unmarshal(msg.Value, &account); err != nil {
				return fmt.Errorf("failed to unmarshal bank user registered: %v", err)
			}

			var status UserStatus
			switch account.Status {
			case pbbank.Account_REGISTERED:
				status = UserStatusBankAccountRegistered
			case pbbank.Account_REJECTED:
				status = UserStatusBankAccountRejected
			case pbbank.Account_NEW:
			default:
				return fmt.Errorf("unknown bank status: %v", account.Status)
			}

			_, err = i.repo.SetUserBank(ctx, account.UserId, status, account.AccountId, account.RejectReason)
			if err != nil {
				return fmt.Errorf("failed to set user bank: %v", err)
			}

			if err := i.bankAccountsReader.CommitMessages(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit message: %v", err)
			}
		}
	}
}

func (i *Interop) NewUserEvent(ctx context.Context, u *User) error {
	pbuser := pb.User{
		UserId: u.UserID,
		Email:  u.Email,
		Status: pb.UserStatus_NEW,
	}

	val, err := proto.Marshal(&pbuser)
	if err != nil {
		log.Fatalln("failed to marshal user:", err)
	}

	if err := i.usersWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(pbuser.UserId),
		Value: val,
	}); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
