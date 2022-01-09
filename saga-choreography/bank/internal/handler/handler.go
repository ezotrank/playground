package handler

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/repository"
	pbwallet "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

type IRepository interface {
	SaveAccount(ctx context.Context, account *repository.Account) error
	SaveTransaction(ctx context.Context, transaction *repository.Transaction) error
}

type IProducer interface {
	NewAccountEvent(ctx context.Context, account *repository.Account) error
	NewTransactionEvent(ctx context.Context, transaction *repository.Transaction) error
}

type IExternalServiceClient interface {
	CreateAccount(ctx context.Context, account *repository.Account) error
}

func NewHandler(repo IRepository, producer IProducer, external IExternalServiceClient) *Handler {
	return &Handler{
		repo:     repo,
		producer: producer,
		external: external,
	}
}

type Handler struct {
	repo     IRepository
	external IExternalServiceClient
	producer IProducer
}

func (h *Handler) WalletUsersHandler(ctx context.Context, msg kafka.Message) error {
	var user pbwallet.User
	if err := proto.Unmarshal(msg.Value, &user); err != nil {
		return fmt.Errorf("failed to unmarshal wallet user: %v", err)
	}

	account := &repository.Account{
		AccountID: user.UserId,
		UserID:    user.UserId,
		Status:    repository.AccountStatusRegistered,
	}

	if err := h.external.CreateAccount(ctx, account); err != nil {
		return fmt.Errorf("failed to create account: %v", err)
	}

	if err := h.repo.SaveAccount(ctx, account); err != nil {
		return fmt.Errorf("failed to save account: %v", err)
	}

	if err := h.producer.NewAccountEvent(ctx, account); err != nil {
		return fmt.Errorf("failed to send account event: %v", err)
	}

	return nil
}

func (h *Handler) WalletTransactionsHandler(ctx context.Context, msg kafka.Message) error {
	var pbtrx pbwallet.Transaction
	if err := proto.Unmarshal(msg.Value, &pbtrx); err != nil {
		return fmt.Errorf("failed to unmarshal wallet transaction: %v", err)
	}

	trx := &repository.Transaction{
		TransactionID: pbtrx.TransactionId,
		AccountID:     pbtrx.UserId, // TODO(ezo): change to account id
		Amount:        int(pbtrx.Amount),
	}

	trx.Status = repository.TransactionStatusSucceed
	if err := h.repo.SaveTransaction(ctx, trx); err != nil {
		return fmt.Errorf("failed to save transaction: %v", err)
	}

	if err := h.producer.NewTransactionEvent(ctx, trx); err != nil {
		return fmt.Errorf("failed to send transaction event: %v", err)
	}

	return nil
}
