package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

type Server struct {
	repo    *Repository
	interop *Interop
}

func (s *Server) UserCreate(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	user, err := s.repo.SetUserIDAndEmail(ctx, req.UserId, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	if err := s.interop.NewUserEvent(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user through interop: %v", err)
	}

Loop:
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("context canceled")
		default:
			user, err = s.repo.GetUserByID(ctx, user.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user: %v", err)
			}

			if user.Status != UserStatusNew {
				break Loop
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	var status pb.UserStatus
	switch user.Status {
	case UserStatusNew:
		status = pb.UserStatus_USER_STATUS_NEW
	case UserStatusBankAccountRegistered:
		status = pb.UserStatus_USER_STATUS_BANK_ACCOUNT_REGISTERED
	case UserStatusBankAccountRejected:
		status = pb.UserStatus_USER_STATUS_BANK_ACCOUNT_REJECTED
	default:
		return nil, fmt.Errorf("failed to convert user status: %v", user.Status)
	}

	return &pb.UserCreateResponse{
		UserId:           user.UserID,
		Email:            user.Email,
		Status:           status,
		BankRejectReason: user.BankRejectReason,
	}, nil
}

func (s *Server) FundIn(ctx context.Context, req *pb.FundInRequest) (*pb.FundInResponse, error) {
	trx, err := s.repo.TransactionCreate(ctx, req.TransactionId, req.UserId, int(req.Amount))
	if err != nil {
		return nil, fmt.Errorf("failed create transaction: %w", err)
	}

	if err := s.interop.NewTransactionEvent(ctx, trx); err != nil {
		return nil, fmt.Errorf("failed to create user through interop: %v", err)
	}

Loop:
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("context canceled")
		default:
			trx, err = s.repo.TransactionGetByID(ctx, trx.TransactionID)
			if err != nil {
				return nil, fmt.Errorf("failed to get transaction: %v", err)
			}

			if trx.Status != TransactionStatusNew {
				break Loop
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	var status pb.TransactionStatus
	switch trx.Status {
	case TransactionStatusNew:
		status = pb.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED
	case TransactionStatusRejected:
		status = pb.TransactionStatus_TRANSACTION_STATUS_REJECTED
	case TransactionStatusSuccess:
		status = pb.TransactionStatus_TRANSACTION_STATUS_SUCCESS
	default:
		return nil, fmt.Errorf("failed to convert transaction status: %v", trx.Status)
	}

	return &pb.FundInResponse{
		UserId:        trx.UserID,
		TransactionId: trx.TransactionID,
		Amount:        int64(trx.Amount),
		Status:        status,
	}, nil
}

func (s *Server) FundOut(_ context.Context, _ *pb.FundOutRequest) (*pb.FundOutResponse, error) {
	return nil, nil
}

func (s *Server) Check(_ context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *Server) Watch(_ *grpc_health_v1.HealthCheckRequest, _ grpc_health_v1.Health_WatchServer) error {
	return nil
}
