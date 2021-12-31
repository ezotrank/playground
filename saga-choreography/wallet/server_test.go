package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	pbbank "github.com/ezotrank/playground/saga-choreography/bank/proto/gen/go/bank/v1"
	pb "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

func TestServer_CreateUser(t *testing.T) {
	defer purge(t)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer func() {
		cancel()
		wg.Wait()
	}()

	brokers := []string{"localhost:" + resources["kafka"].GetPort("9094/tcp")}

	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topicWalletUsers,
		GroupID: consumerGroup,
	})

	producer := &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: topicBankAccounts,
	}

	require.NoError(t, CreateTopic(brokers, topicBankAccounts))

	msgcount := 0
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := consumer.FetchMessage(context.Background())
				require.NoError(t, err)

				var user pb.User
				require.NoError(t, proto.Unmarshal(msg.Value, &user))
				require.Equal(t, user.UserId, "1")
				require.Equal(t, user.Email, "user@example.com")

				val, err := proto.Marshal(&pbbank.Account{
					AccountId: "1",
					UserId:    user.UserId,
					Status:    pbbank.Account_STATUS_REGISTERED,
				})
				require.NoError(t, err)

				require.NoError(t, producer.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte("1"),
					Value: val,
				}))

				// The numbers of messages sent by interop
				msgcount++
			}
		}
	}()

	repo := &Repository{
		rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:" + resources["redis"].GetPort("6379/tcp"),
		}),
	}

	interop, err := NewInterop(brokers, repo)
	require.NoError(t, err)

	wg.Add(1)
	go func() {
		wg.Done()
		require.NoError(t, interop.Start(ctx))
	}()

	client, err := getService(ctx, &Server{
		repo:    repo,
		interop: interop,
	})
	require.NoError(t, err)

	{
		resp, err := client.UserCreate(ctx, &pb.UserCreateRequest{
			UserId: "1",
			Email:  "user@example.com",
		})
		require.NoError(t, err)

		want := pb.UserCreateResponse{
			UserId: "1",
			Email:  "user@example.com",
			Status: pb.UserStatus_USER_STATUS_BANK_ACCOUNT_REGISTERED,
		}
		require.Equal(t, want.String(), resp.String())
	}

	{
		// Check idempotency in rpc
		resp, err := client.UserCreate(ctx, &pb.UserCreateRequest{
			UserId: "1",
			Email:  "user@example.com",
		})
		require.NoError(t, err)

		want := pb.UserCreateResponse{
			UserId: "1",
			Email:  "user@example.com",
			Status: pb.UserStatus_USER_STATUS_BANK_ACCOUNT_REGISTERED,
		}
		require.Equal(t, want.String(), resp.String())
	}

	// Check that we send two messages to the bank topic
	require.Eventually(t, func() bool {
		return msgcount == 2
	}, 5*time.Second, 100*time.Millisecond)
}
