package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type UserStatus string

const (
	UserStatusNew                   UserStatus = "NEW"
	UserStatusBankAccountRegistered UserStatus = "BANK_ACCOUNT_REGISTERED"
	UserStatusBankAccountRejected   UserStatus = "BANK_ACCOUNT_REJECTED"
)

type User struct {
	UserID string     `json:"id"`
	Email  string     `json:"email"`
	Status UserStatus `json:"bank_status"`

	BankAccountID    string `json:"bank_account"`
	BankRejectReason string `json:"bank_reject_reason"`
}

type Repository struct {
	rdb *redis.Client
}

var ErrNotFound = errors.New("user not found")

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	cmd := r.rdb.Get(ctx, id)
	if cmd.Err() == redis.Nil {
		return nil, ErrNotFound
	} else if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to get user: %v", cmd.Err())
	}

	var user User
	if err := json.Unmarshal([]byte(cmd.Val()), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %v", err)
	}

	return &user, nil
}

func (r *Repository) SetUserIDAndEmail(ctx context.Context, id, email string) (*User, error) {
	user, err := r.GetUserByID(ctx, id)
	if err != nil && err == ErrNotFound {
		user = &User{
			UserID: id,
			Email:  email,
			Status: UserStatusNew,
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	val, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %v", err)
	}
	cmd := r.rdb.Set(ctx, user.UserID, val, 0)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to set user: %v", cmd.Err())
	}

	return user, nil
}

func (r *Repository) SetUserBank(ctx context.Context, id string, status UserStatus, account, reason string) (*User, error) {
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	user.Status = status
	user.BankAccountID = account
	user.BankRejectReason = reason

	val, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %v", err)
	}
	cmd := r.rdb.Set(ctx, user.UserID, val, 0)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to set user bank: %v", cmd.Err())
	}

	return user, nil
}

type TransactionStatus string
type TransactionType string

const (
	TransactionStatusNew      TransactionStatus = "NEW"
	TransactionStatusSuccess  TransactionStatus = "SUCCESS"
	TransactionStatusRejected TransactionStatus = "REJECTED"
)

type Transaction struct {
	TransactionID  string
	UserID         string
	Amount         int
	Status         TransactionStatus
	RejectedReason string
}

func (r *Repository) TransactionCreate(ctx context.Context, trxid, uid string, amount int) (*Transaction, error) {
	trx, err := r.TransactionGetByID(ctx, trxid)
	if err != nil && err == ErrNotFound {
		trx = &Transaction{
			TransactionID: trxid,
			UserID:        uid,
			Amount:        amount,
			Status:        TransactionStatusNew,
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}

	val, err := json.Marshal(trx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %v", err)
	}
	cmd := r.rdb.Set(ctx, trx.TransactionID, val, 0)
	if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to set user bank: %v", cmd.Err())
	}

	return trx, nil
}

func (r *Repository) TransactionGetByID(ctx context.Context, trxid string) (*Transaction, error) {
	cmd := r.rdb.Get(ctx, trxid)
	if cmd.Err() == redis.Nil {
		return nil, ErrNotFound
	} else if cmd.Err() != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", cmd.Err())
	}

	var trx Transaction
	if err := json.Unmarshal([]byte(cmd.Val()), &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %v", err)
	}

	return &trx, nil
}

func (r *Repository) TransactionUpdateStatus(ctx context.Context, trxid string, status TransactionStatus, reason string) error {
	trx, err := r.TransactionGetByID(ctx, trxid)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %v", err)
	}

	trx.Status = status
	trx.RejectedReason = reason

	val, err := json.Marshal(trx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}
	cmd := r.rdb.Set(ctx, trx.TransactionID, val, 0)
	if cmd.Err() != nil {
		return fmt.Errorf("failed to update transaction: %v", cmd.Err())
	}

	return nil
}
