package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/external"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/handler"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/interop"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/producer"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/repository"
	pb "github.com/ezotrank/playground/saga-choreography/bank/proto/gen/go/bank/v1"
	pbwallet "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

func TestWalletUsers(t *testing.T) {
	tests := []struct {
		name            string
		walletEvent     kafka.Message
		account         repository.Account
		bankEvent       kafka.Message
		externalHandler http.HandlerFunc

		wantErr bool
	}{
		{
			name: "success case",
			walletEvent: kafka.Message{
				Topic: "wallet.users",
				Key:   []byte("user"),
				Value: marshal(&pbwallet.User{
					UserId: "user",
					Email:  "user@example.com",
					Status: pbwallet.UserStatus_USER_STATUS_NEW,
				}),
			},
			account: repository.Account{
				AccountID: "user",
				UserID:    "user",
				Status:    repository.AccountStatusRegistered,
			},
			bankEvent: kafka.Message{
				Topic: "bank.accounts",
				Key:   []byte("user"),
				Value: marshal(&pb.Account{
					AccountId: "user",
					UserId:    "user",
					Status:    pb.Account_STATUS_REGISTERED,
				}),
			},
			externalHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broker := "localhost:" + resources["kafka"].GetPort("9094/tcp")
			require.NoError(t, CreateTopic([]string{broker}, "wallet.users"))
			require.NoError(t, CreateTopic([]string{broker}, "bank.accounts"))

			writer := &kafka.Writer{
				Addr: kafka.TCP(broker),
			}
			require.NoError(t, writer.WriteMessages(context.Background(), tt.walletEvent))

			ts := httptest.NewServer(tt.externalHandler)
			defer ts.Close()

			hndlr := handler.NewHandler(
				repository.NewRepository(
					"localhost:"+resources["redis"].GetPort("6379/tcp"),
					0,
				),
				producer.NewProducer(broker),
				external.NewExternal(ts.URL),
			)
			intr, err := interop.NewInterop([]string{broker}, interop.Flow{
				Rules: map[string]interop.Rule{
					"wallet.users": {
						Handler:  hndlr.WalletUsersHandler,
						Attempts: 1,
					},
				},
			})
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: []string{broker},
				GroupID: "test",
				Topic:   "bank.accounts",
			})

			go func() {
				defer cancel()
				msg, err := reader.FetchMessage(ctx)
				require.NoError(t, err)
				require.Equal(t, tt.bankEvent.Topic, msg.Topic)
				require.Equal(t, tt.bankEvent.Key, msg.Key)
				require.Equal(t, tt.bankEvent.Value, msg.Value)
			}()

			require.NoError(t, intr.Start(ctx))
		})
	}
}

func marshal(msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return data
}
