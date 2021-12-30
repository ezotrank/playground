package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/ezotrank/playground/saga-choreography/wallet/gen"
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
	if val, ok := pb.UserStatus_value[string(user.Status)]; ok {
		status = pb.UserStatus(val)
	} else {
		return nil, fmt.Errorf("failed to convert user status: %v", user.Status)
	}

	return &pb.UserCreateResponse{
		UserId:           user.UserID,
		Email:            user.Email,
		Status:           status,
		BankRejectReason: user.BankRejectReason,
	}, nil
}

func (s *Server) FundIn(_ context.Context, _ *pb.FundInRequest) (*pb.FundInResponse, error) {
	return nil, nil
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
