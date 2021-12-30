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

var ErrUserNotFound = errors.New("user not found")

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	cmd := r.rdb.Get(ctx, id)
	if cmd.Err() == redis.Nil {
		return nil, ErrUserNotFound
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
	if err != nil && err == ErrUserNotFound {
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
