package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"

	pb "github.com/ezotrank/playground/saga-choreography/bank/proto/gen/go/bank/v1"
)

func NewProducer(brokers ...string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

type Producer struct {
	writer *kafka.Writer
}

func (p *Producer) NewAccountEvent(ctx context.Context, account *Account) error {
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

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(pbaccount.AccountId),
		Value: val,
		Topic: topicBankAccounts,
	}); err != nil {
		return fmt.Errorf("failed to write message: %p", err)
	}

	return nil
}

func (p *Producer) NewTransactionEvent(ctx context.Context, trx *Transaction) error {
	return nil
}
