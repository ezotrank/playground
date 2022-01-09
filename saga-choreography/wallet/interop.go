package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	pbbank "github.com/ezotrank/playground/saga-choreography/bank/proto/gen/go/bank/v1"
	pb "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

const (
	// producer
	topicWalletUsers        = "wallet.users"
	topicWalletTransactions = "wallet.transactions"

	// consumer
	topicBankAccounts     = "bank.accounts"
	topicBankTransactions = "bank.transactions"
)

const consumerGroup = "wallet-interop"

func NewInterop(brokers []string, repo *Repository) (*Interop, error) {
	if err := CreateTopic(brokers, topicWalletUsers, topicWalletTransactions); err != nil {
		return nil, fmt.Errorf("failed to create topics: %w", err)
	}

	return &Interop{
		repo: repo,
		usersWriter: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topicWalletUsers,
		},
		transactionsWriter: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topicWalletTransactions,
		},
		bankAccountsReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topicBankAccounts,
			GroupID: consumerGroup,
		}),
		bankTransactionsReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topicBankTransactions,
			GroupID: consumerGroup,
		}),
	}, nil
}

func CreateTopic(brokers []string, topics ...string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker: %v", err)
	}

	// Hack to not get a "Not Available: the cluster is in the middle" error in WriteMessages.
	// Create or check if topics exist.
	for _, topic := range topics {
		if err := conn.CreateTopics(
			kafka.TopicConfig{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		); err != nil {
			return fmt.Errorf("failed to create topic: %v", err)
		}
	}

	return nil
}

type Interop struct {
	repo *Repository

	usersWriter        *kafka.Writer
	transactionsWriter *kafka.Writer

	bankAccountsReader     *kafka.Reader
	bankTransactionsReader *kafka.Reader
}

func (i *Interop) Start(ctx context.Context) error {
	errc := make(chan error, 1)

	go func() {
		if err := i.bankAccountsConsumer(ctx); err != nil {
			errc <- err
		}
	}()

	go func() {
		if err := i.bankTransactionsConsumer(ctx); err != nil {
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

func (i *Interop) bankAccountsConsumer(ctx context.Context) error {
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
			case pbbank.Account_STATUS_REGISTERED:
				status = UserStatusBankAccountRegistered
			case pbbank.Account_STATUS_REJECTED:
				status = UserStatusBankAccountRejected
			case pbbank.Account_STATUS_NEW:
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

func (i *Interop) bankTransactionsConsumer(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			ctx := context.Background()
			msg, err := i.bankTransactionsReader.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch message: %v", err)
			}

			var trx pbbank.Transaction
			if err := proto.Unmarshal(msg.Value, &trx); err != nil {
				return fmt.Errorf("failed to unmarshal bank user registered: %v", err)
			}

			var status TransactionStatus
			switch trx.Status {
			case pbbank.Transaction_STATUS_SUCCEED:
				status = TransactionStatusSuccess
			case pbbank.Transaction_STATUS_REJECTED:
				status = TransactionStatusRejected
			default:
				return fmt.Errorf("unknown bank status: %v", trx.Status)
			}

			err = i.repo.TransactionUpdateStatus(ctx, trx.TransactionId, status, trx.RejectReason)
			if err != nil {
				return fmt.Errorf("failed to update transaction: %w", err)
			}

			if err := i.bankTransactionsReader.CommitMessages(ctx, msg); err != nil {
				return fmt.Errorf("failed to commit message: %v", err)
			}
		}
	}
}

func (i *Interop) NewUserEvent(ctx context.Context, u *User) error {
	pbuser := pb.User{
		UserId: u.UserID,
		Email:  u.Email,
		Status: pb.UserStatus_USER_STATUS_NEW,
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

func (i *Interop) NewTransactionEvent(ctx context.Context, trx *Transaction) error {
	pbtrx := pb.Transaction{
		TransactionId: trx.TransactionID,
		UserId:        trx.UserID,
		Amount:        int64(trx.Amount),
		// TODO(ezo): should be empty or need to map?
		Status: pb.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED,
	}

	val, err := proto.Marshal(&pbtrx)
	if err != nil {
		log.Fatalln("failed to marshal user:", err)
	}

	if err := i.transactionsWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(pbtrx.UserId),
		Value: val,
	}); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
