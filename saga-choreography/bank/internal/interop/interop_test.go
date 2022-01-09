//go:generate mockgen -source=interop.go -destination=../mocks/mock_interop.go -package=mocks

package interop

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/mocks"
)

func TestInterop_Start(t *testing.T) {
	type fields struct {
		flow   Flow
		reader *mocks.Mockireader
		writer *mocks.Mockiwriter
	}
	tests := []struct {
		name    string
		flow    Flow
		prepare func(f *fields)
		wantErr bool
	}{
		{
			name: "success flow",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return nil
						},
					},
				},
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						DoAndReturn(func(_ context.Context) (kafka.Message, error) {
							return kafka.Message{}, io.EOF
						}),
				)
				f.reader.EXPECT().
					CommitMessages(gomock.Any(), kafka.Message{
						Topic: "topic1",
						Headers: []kafka.Header{
							{attemptsHeader, []byte("1")},
						},
					}).
					Return(nil).
					Times(1)
			},
			wantErr: false,
		},
		{
			name: "fetch message error",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return nil
						},
					},
				},
			},
			prepare: func(f *fields) {
				f.reader.EXPECT().
					FetchMessage(gomock.Any()).
					Return(kafka.Message{
						Topic: "topic1",
					}, fmt.Errorf("error"))
			},
			wantErr: true,
		},
		{
			name: "message with unknown topic",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return nil
						},
					},
				},
			},
			prepare: func(f *fields) {
				f.reader.EXPECT().
					FetchMessage(gomock.Any()).
					Return(kafka.Message{
						Topic: "topic2",
					}, nil)
			},
			wantErr: true,
		},
		{
			name: "handle return error without retry policy",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return fmt.Errorf("error")
						},
						Attempts: 1,
					},
				},
			},
			prepare: func(f *fields) {
				f.reader.EXPECT().
					FetchMessage(gomock.Any()).
					Return(kafka.Message{
						Topic: "topic1",
					}, nil)
			},
			wantErr: true,
		},
		{
			name: "commit message return error",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return nil
						},
						Attempts: 1,
					},
				},
			},
			prepare: func(f *fields) {
				f.reader.EXPECT().
					FetchMessage(gomock.Any()).
					Return(kafka.Message{
						Topic: "topic1",
					}, nil)
				f.reader.EXPECT().
					CommitMessages(gomock.Any(), kafka.Message{
						Topic: "topic1",
						Headers: []kafka.Header{
							{attemptsHeader, []byte("1")},
						},
					}).
					Return(fmt.Errorf("error"))
			},
			wantErr: true,
		},
		{
			name: "handle return error with retry policy",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return fmt.Errorf("error")
						},
						Attempts: 3,
					},
				},
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}, nil),
				)
				gomock.InOrder(
					f.writer.EXPECT().
						WriteMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}).
						Return(nil),
					f.writer.EXPECT().
						WriteMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}).
						Return(nil),
				)
				gomock.InOrder(
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}).
						Return(nil),
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}).
						Return(nil),
				)
			},
			wantErr: true,
		},
		{
			name: "handle return error with retry policy and DLQ",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return fmt.Errorf("error")
						},
						DLQ: "dlq",
					},
				},
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						DoAndReturn(func(_ context.Context) (kafka.Message, error) {
							return kafka.Message{}, io.EOF
						}),
				)
				f.writer.EXPECT().
					WriteMessages(gomock.Any(), kafka.Message{
						Topic: "dlq",
						Headers: []kafka.Header{
							{attemptsHeader, []byte("0")},
						},
					}).
					Return(nil)
				f.reader.EXPECT().
					CommitMessages(gomock.Any(), kafka.Message{
						Topic: "topic1",
						Headers: []kafka.Header{
							{attemptsHeader, []byte("1")},
						},
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "handle return error only first retry policy is set",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							if getAttempts(msg.Headers) == 2 {
								return nil
							}
							return fmt.Errorf("error")
						},
						Attempts: 2,
					},
				},
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						DoAndReturn(func(_ context.Context) (kafka.Message, error) {
							return kafka.Message{}, io.EOF
						}),
				)
				f.writer.EXPECT().
					WriteMessages(gomock.Any(), kafka.Message{
						Topic: "topic1",
						Headers: []kafka.Header{
							{attemptsHeader, []byte("1")},
						},
					}).
					Return(nil)
				gomock.InOrder(
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}).
						Return(nil),
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}).
						Return(nil),
				)
			},
			wantErr: false,
		},
		{
			name: "handle return error dlq is set with retries",
			flow: Flow{
				Rules: map[string]Rule{
					"topic1": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return fmt.Errorf("error")
						},
						Attempts: 1,
						DLQ:      "retry",
					},
					"retry": {
						Handler: func(ctx context.Context, msg kafka.Message) error {
							return fmt.Errorf("error")
						},
						Attempts: 3,
						DLQ:      "dlq",
					},
				},
			},
			prepare: func(f *fields) {
				gomock.InOrder(
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "topic1",
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("0")},
							},
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						Return(kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}, nil),
					f.reader.EXPECT().
						FetchMessage(gomock.Any()).
						DoAndReturn(func(_ context.Context) (kafka.Message, error) {
							return kafka.Message{}, io.EOF
						}),
				)
				gomock.InOrder(
					f.writer.EXPECT().
						WriteMessages(gomock.Any(), kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("0")},
							},
						}).
						Return(nil),
					f.writer.EXPECT().
						WriteMessages(gomock.Any(), kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}).
						Return(nil),
					f.writer.EXPECT().
						WriteMessages(gomock.Any(), kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}).
						Return(nil),
					f.writer.EXPECT().
						WriteMessages(gomock.Any(), kafka.Message{
							Topic: "dlq",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("0")},
							},
						}).
						Return(nil),
				)
				gomock.InOrder(
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "topic1",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}).
						Return(nil),
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("1")},
							},
						}).
						Return(nil),
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("2")},
							},
						}).
						Return(nil),
					f.reader.EXPECT().
						CommitMessages(gomock.Any(), kafka.Message{
							Topic: "retry",
							Headers: []kafka.Header{
								{attemptsHeader, []byte("3")},
							},
						}).
						Return(nil),
				)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			f := fields{
				flow:   tt.flow,
				reader: mocks.NewMockireader(ctrl),
				writer: mocks.NewMockiwriter(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			f.reader.EXPECT().Close().Return(nil)
			f.writer.EXPECT().Close().Return(nil)

			i := &Interop{
				flow:   f.flow,
				reader: f.reader,
				writer: f.writer,
			}

			if err := i.Start(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
