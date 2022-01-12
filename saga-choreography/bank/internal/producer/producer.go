package producer

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/repository"
	pb "github.com/ezotrank/playground/saga-choreography/bank/proto/gen/go/bank/v1"
)

type Topics struct {
	BankAccountsTopic string
}

func NewProducer(topics Topics, brokers ...string) *Producer {
	return &Producer{
		BankAccountsTopic: topics.BankAccountsTopic,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

type Producer struct {
	BankAccountsTopic string

	writer *kafka.Writer
}

func (p *Producer) NewAccountEvent(ctx context.Context, account *repository.Account) error {
	var status pb.Account_Status
	switch account.Status {
	case repository.AccountStatusRegistered:
		status = pb.Account_STATUS_REGISTERED
	case repository.AccountStatusRejected:
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

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(pbaccount.AccountId),
		Value: val,
		Topic: p.BankAccountsTopic,
	}); err != nil {
		return fmt.Errorf("failed to write message: %p", err)
	}

	return nil
}

func (p *Producer) NewTransactionEvent(ctx context.Context, trx *repository.Transaction) error {
	return nil
}
