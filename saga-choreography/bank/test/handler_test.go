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
	broker := KafkaGetBroker()
	redis := RedisAddr()

	type fields struct {
		topicu string
		topica string
		gc     string
		repo   *repository.Repository
		flow   interop.Flow
	}
	tests := []struct {
		name      string
		prepare   func(f *fields)
		accounts  []*repository.Account
		umsg      []kafka.Message
		amsg      []kafka.Message
		topicuoff int64
		wantErr   bool
	}{
		{
			name: "success case",
			prepare: func(f *fields) {
				f.topicu = randstr()
				f.topica = randstr()
				f.gc = randstr()

				require.NoError(t, KafkaCreateTopic([]string{broker}, f.topicu, f.topica))

				f.repo = repository.NewRepository(redis, 0)

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				hndlr := handler.NewHandler(
					f.repo,
					producer.NewProducer(producer.Topics{
						BankAccountsTopic: f.topica,
					}, broker),
					external.NewExternal(srv.URL),
				)

				f.flow = interop.Flow{
					Rules: map[string]interop.Rule{
						f.topicu: {
							Handler:  hndlr.WalletUsersHandler,
							Attempts: 1,
						},
					},
				}
			},
			accounts: []*repository.Account{
				{
					AccountID: "bob",
					UserID:    "bob",
					Status:    repository.AccountStatusRegistered,
				},
				{
					AccountID: "alice",
					UserID:    "alice",
					Status:    repository.AccountStatusRegistered,
				},
			},
			umsg: []kafka.Message{
				{
					Key: []byte("bob"),
					Value: marshal(&pbwallet.User{
						UserId: "bob",
						Email:  "bob@example.com",
						Status: pbwallet.UserStatus_USER_STATUS_NEW,
					}),
				},
				{
					Key: []byte("alice"),
					Value: marshal(&pbwallet.User{
						UserId: "alice",
						Email:  "alice@example.com",
						Status: pbwallet.UserStatus_USER_STATUS_NEW,
					}),
				},
			},
			amsg: []kafka.Message{
				{
					Key: []byte("bob"),
					Value: marshal(&pb.Account{
						AccountId: "bob",
						UserId:    "bob",
						Status:    pb.Account_STATUS_REGISTERED,
					}),
				},
				{
					Key: []byte("alice"),
					Value: marshal(&pb.Account{
						AccountId: "alice",
						UserId:    "alice",
						Status:    pb.Account_STATUS_REGISTERED,
					}),
				},
			},
			topicuoff: 2,
			wantErr:   false,
		},
		{
			name: "external service returns error",
			prepare: func(f *fields) {
				f.topicu = randstr()
				f.topica = randstr()
				f.gc = randstr()

				require.NoError(t, KafkaCreateTopic([]string{broker}, f.topicu, f.topica))

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))

				hndlr := handler.NewHandler(
					&repository.Repository{},
					&producer.Producer{},
					external.NewExternal(srv.URL),
				)

				f.flow = interop.Flow{
					Rules: map[string]interop.Rule{
						f.topicu: {
							Handler:  hndlr.WalletUsersHandler,
							Attempts: 1,
						},
					},
				}
			},
			accounts: []*repository.Account{},
			umsg: []kafka.Message{
				{
					Key: []byte("alice"),
					Value: marshal(&pbwallet.User{
						UserId: "alice",
						Email:  "alice@example.com",
						Status: pbwallet.UserStatus_USER_STATUS_NEW,
					}),
				},
			},
			amsg:      []kafka.Message{},
			topicuoff: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer purge()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			f := fields{}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			intr, err := interop.NewInterop([]string{broker}, f.flow, f.gc)
			require.NoError(t, err)
			done := make(chan struct{})
			go func() {
				require.Equal(t, tt.wantErr, intr.Start(ctx) != nil)
				done <- struct{}{}
			}()

			writer := &kafka.Writer{
				Addr:  kafka.TCP(broker),
				Topic: f.topicu,
			}
			for _, msg := range tt.umsg {
				require.NoError(t, writer.WriteMessages(ctx, msg))
			}

			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: []string{broker},
				Topic:   f.topica,
			})

			for _, want := range tt.amsg {
				msg, err := reader.FetchMessage(ctx)
				require.NoError(t, err)
				require.Equal(t, want.Key, msg.Key)
				require.Equal(t, want.Value, msg.Value)
			}
			time.Sleep(5 * time.Second) // Can be raise condition, when ctx is closed but intr.Start hasn't started yet.
			cancel()
			<-done

			client := kafka.Client{
				Addr: kafka.TCP(broker),
			}
			resp, err := client.OffsetFetch(context.Background(), &kafka.OffsetFetchRequest{
				Addr:    kafka.TCP(broker),
				GroupID: f.gc,
				Topics: map[string][]int{
					f.topicu: {0},
				},
			})
			require.NoError(t, err)
			require.NoError(t, resp.Error)
			require.Equal(t, tt.topicuoff, resp.Topics[f.topicu][0].CommittedOffset)

			for _, want := range tt.accounts {
				got, err := f.repo.GetAccountByID(context.Background(), want.AccountID)
				require.NoError(t, err)
				require.Equal(t, want, got)
			}
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
