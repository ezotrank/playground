package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type AccountStatus string
type TransactionStatus string

const (
	AccountStatusRegistered AccountStatus = "REGISTERED"
	AccountStatusRejected   AccountStatus = "REJECTED"

	TransactionStatusSucceed TransactionStatus = "SUCCEED"
	TransactionStatusFailed  TransactionStatus = "FAILED"
)

type Account struct {
	AccountID    string        `json:"account_id"`
	UserID       string        `json:"user_id"`
	Status       AccountStatus `json:"status"`
	RejectReason string        `json:"reject_reason"`
}

type Transaction struct {
	TransactionID string            `json:"transaction_id"`
	AccountID     string            `json:"account_id"`
	Amount        int               `json:"amount"`
	Status        TransactionStatus `json:"status"`
	FailedReason  string            `json:"failed_reason"`
}

func NewRepository(addr string, db int) *Repository {
	return &Repository{
		rdb: redis.NewClient(&redis.Options{
			Addr: addr,
			DB:   db,
		}),
	}
}

type Repository struct {
	rdb *redis.Client
}

func (r *Repository) SaveAccount(ctx context.Context, account *Account) error {
	val, err := json.Marshal(&account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}
	if err := r.rdb.Set(ctx, account.AccountID, val, 0).Err(); err != nil {
		return fmt.Errorf("failed to save account: %w", err)
	}
	return nil
}

func (r *Repository) SaveTransaction(ctx context.Context, trx *Transaction) error {
	val, err := json.Marshal(&trx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}
	if err := r.rdb.Set(ctx, trx.TransactionID, val, 0).Err(); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}
	return nil
}

func (r *Repository) AccountGetByUserID(ctx context.Context, aid string) (*Account, error) {
	return nil, nil
}

func (r *Repository) GetAccountByID(ctx context.Context, aid string) (*Account, error) {
	val, err := r.rdb.Get(ctx, aid).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var account Account
	if err := json.Unmarshal([]byte(val), &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}
	return &account, nil
}