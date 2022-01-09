//go:generate mockgen -source=handler.go -destination=../mocks/mock_handler.go -package=mocks

package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/mocks"
	"github.com/ezotrank/playground/saga-choreography/bank/internal/repository"
	pbwallet "github.com/ezotrank/playground/saga-choreography/wallet/proto/gen/go/wallet/v1"
)

func marshal(msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return data
}

func TestHandler_WalletUsersHandler(t *testing.T) {
	type fields struct {
		repo     *mocks.MockIRepository
		external *mocks.MockIExternalServiceClient
		producer *mocks.MockIProducer
	}
	type args struct {
		ctx context.Context
		msg kafka.Message
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		wantErr bool
	}{
		{
			name: "success event processing",
			prepare: func(f *fields) {
				f.external.EXPECT().
					CreateAccount(gomock.Any(), &repository.Account{
						AccountID: "1",
						UserID:    "1",
						Status:    repository.AccountStatusRegistered,
					}).
					Return(nil).
					Times(1)
				f.repo.
					EXPECT().
					SaveAccount(gomock.Any(), &repository.Account{
						AccountID: "1",
						UserID:    "1",
						Status:    repository.AccountStatusRegistered,
					}).
					Return(nil).
					Times(1)
				f.producer.
					EXPECT().
					NewAccountEvent(gomock.Any(), &repository.Account{
						AccountID: "1",
						UserID:    "1",
						Status:    repository.AccountStatusRegistered,
					}).
					Return(nil).
					Times(1)
			},
			args: args{
				ctx: context.Background(),
				msg: kafka.Message{
					Value: marshal(&pbwallet.User{
						UserId: "1",
						Email:  "user@example.com",
					}),
				},
			},
			wantErr: false,
		},
		{
			name: "failed to send new account event",
			prepare: func(f *fields) {
				f.external.EXPECT().
					CreateAccount(gomock.Any(), &repository.Account{
						AccountID: "1",
						UserID:    "1",
						Status:    repository.AccountStatusRegistered,
					}).
					Return(nil).
					Times(1)
				f.repo.
					EXPECT().
					SaveAccount(gomock.Any(), &repository.Account{
						AccountID: "1",
						UserID:    "1",
						Status:    repository.AccountStatusRegistered,
					}).
					Return(nil).
					Times(1)
				f.producer.
					EXPECT().
					NewAccountEvent(gomock.Any(), &repository.Account{
						AccountID: "1",
						UserID:    "1",
						Status:    repository.AccountStatusRegistered,
					}).
					Return(fmt.Errorf("failed to send")).
					Times(1)
			},
			args: args{
				ctx: context.Background(),
				msg: kafka.Message{
					Value: marshal(&pbwallet.User{
						UserId: "1",
						Email:  "user@example.com",
					}),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			f := fields{
				repo:     mocks.NewMockIRepository(ctrl),
				external: mocks.NewMockIExternalServiceClient(ctrl),
				producer: mocks.NewMockIProducer(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			h := &Handler{
				repo:     f.repo,
				external: f.external,
				producer: f.producer,
			}

			if err := h.WalletUsersHandler(tt.args.ctx, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("WalletUsersHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
