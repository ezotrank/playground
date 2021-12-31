package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	pbwallet "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

func TestUserRegistrationFlow(t *testing.T) {
	walletAddr := "localhost:" + resources["wallet"].GetPort("8080/tcp")

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}
	conn, err := grpc.DialContext(context.Background(), walletAddr, opts...)
	require.NoError(t, err)

	client := pbwallet.NewWalletServiceClient(conn)
	resp, err := client.UserCreate(context.Background(), &pbwallet.UserCreateRequest{
		UserId: "1",
		Email:  "user@example.com",
	})
	require.NoError(t, err)

	want := pbwallet.UserCreateResponse{
		UserId: "1",
		Email:  "user@example.com",
		Status: pbwallet.UserStatus_USER_STATUS_BANK_ACCOUNT_REGISTERED,
	}
	require.Equal(t, want.String(), resp.String())
}
