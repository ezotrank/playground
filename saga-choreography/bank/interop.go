package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	pb "github.com/ezotrank/playground/saga-choreography/bank/proto/gen/go/bank/v1"
	pbwallet "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

const (
	// producer
	topicBankAccounts = "bank.accounts"

	// consumer
	topicWalletUsers = "wallet.users"
)

const consumerGroup = "bank-interop"

func NewInterop(brokers []string, repo *Repository) (*Interop, error) {
	if err := CreateTopic(brokers, topicBankAccounts); err != nil {
		return nil, fmt.Errorf("failed to create topic %s: %w", topicBankAccounts, err)
	}

	return &Interop{
		repo: repo,
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topicBankAccounts,
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topicWalletUsers,
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
	repo   *Repository
	writer *kafka.Writer
	reader *kafka.Reader
}

func (i *Interop) Start(ctx context.Context) error {
	errc := make(chan error, 1)

	go func() {
		if err := i.userConsumer(ctx); err != nil {
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
		return i.reader.Close()
	})
	g.Go(func() error {
		return i.writer.Close()
	})

	return g.Wait()
}

func (i *Interop) userConsumer(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			ctx := context.Background()
			msg, err := i.reader.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch message: %v", err)
			}

			var user pbwallet.User
			if err := proto.Unmarshal(msg.Value, &user); err != nil {
				return fmt.Errorf("failed to unmarshal wallet user: %v", err)
			}

			account := &Account{
				AccountID: user.UserId,
				UserID:    user.UserId,
				Status:    AccountStatusRegistered,
			}

			if strings.HasPrefix(user.UserId, "bad") {
				account.Status = AccountStatusRejected
				account.RejectReason = "bad user id"
			}

			if _, err := i.repo.SaveAccount(ctx, account); err != nil {
				return fmt.Errorf("failed to save account: %v", err)
			}

			if err := i.NewAccountEvent(ctx, account); err != nil {
				return fmt.Errorf("failed to send account event: %v", err)
			}

			if err := i.reader.CommitMessages(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit message: %v", err)
			}
		}
	}
}

func (i *Interop) NewAccountEvent(ctx context.Context, account *Account) error {
	var status pb.Account_Status
	switch account.Status {
	case AccountStatusRegistered:
		status = pb.Account_STATUS_REGISTERED
	case AccountStatusRejected:
		status = pb.Account_STATUS_REJECTED
	default:
		return fmt.Errorf("invalid account status: %s", account.Status)
	}

	pbaccount := pb.Account{
		AccountId: account.AccountID,
		UserId:    account.UserID,
		Status:    status,
	}

	val, err := proto.Marshal(&pbaccount)
	if err != nil {
		log.Fatalln("failed to marshal account:", err)
	}

	if err := i.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(pbaccount.AccountId),
		Value: val,
	}); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
