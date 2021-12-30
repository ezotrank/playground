package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type AccountStatus string

const (
	AccountStatusRegistered AccountStatus = "REGISTERED"
	AccountStatusRejected   AccountStatus = "REJECTED"
)

type Account struct {
	AccountID    string        `json:"account_id"`
	UserID       string        `json:"user_id"`
	Status       AccountStatus `json:"status"`
	RejectReason string        `json:"reject_reason"`
}

type Repository struct {
	rdb *redis.Client
}

func (r *Repository) SaveAccount(ctx context.Context, account *Account) (*Account, error) {
	val, err := json.Marshal(&account)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal account: %w", err)
	}
	if err := r.rdb.Set(ctx, account.AccountID, val, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}
	return account, nil
}
