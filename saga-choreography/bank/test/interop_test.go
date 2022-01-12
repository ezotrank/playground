package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/interop"
)

func TestInterop(t *testing.T) {
	broker := KafkaGetBroker()

	type fields struct {
		topic       string
		topicr      string
		topicd      string
		gc          string
		flow        interop.Flow
		handlerRecv chan kafka.Message
	}
	tests := []struct {
		name       string
		prepare    func(f *fields)
		msgs       []kafka.Message
		topicmsg   []kafka.Message
		topicrmsg  []kafka.Message
		topicdmsg  []kafka.Message
		handlermsg []kafka.Message
		topicoff   int64
		topicroff  int64
		topicdoff  int64
		wantErr    bool
	}{
		{
			name: "success case",
			prepare: func(f *fields) {
				f.topic = randstr()
				f.topicr = randstr()
				f.topicd = randstr()
				f.gc = randstr()

				require.NoError(t, KafkaCreateTopic([]string{broker}, f.topic, f.topicr, f.topicd))

				f.handlerRecv = make(chan kafka.Message)
				f.flow = interop.Flow{
					Rules: map[string]interop.Rule{
						f.topic: {
							Handler: func(ctx context.Context, msg kafka.Message) error {
								f.handlerRecv <- msg
								return fmt.Errorf("error")
							},
							Attempts: 1,
							DLQ:      f.topicr,
						},
						f.topicr: {
							Handler: func(ctx context.Context, msg kafka.Message) error {
								f.handlerRecv <- msg
								return fmt.Errorf("error")
							},
							Attempts: 2,
							DLQ:      f.topicd,
						},
						f.topicd: {
							Handler: func(ctx context.Context, msg kafka.Message) error {
								f.handlerRecv <- msg
								return nil
							},
							Attempts: 1,
						},
					},
				}
			},
			msgs: []kafka.Message{
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
			},
			handlermsg: []kafka.Message{
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
			},
			topicrmsg: []kafka.Message{
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
			},
			topicdmsg: []kafka.Message{
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
			},
			topicoff:  1,
			topicroff: 2,
			topicdoff: 1,
			wantErr:   false,
		},
		{
			name: "failure case handler returns error",
			prepare: func(f *fields) {
				f.topic = randstr()
				f.topicr = randstr()
				f.topicd = randstr()
				f.gc = randstr()

				require.NoError(t, KafkaCreateTopic([]string{broker}, f.topic, f.topicr, f.topicd))

				f.handlerRecv = make(chan kafka.Message)
				f.flow = interop.Flow{
					Rules: map[string]interop.Rule{
						f.topic: {
							Handler: func(ctx context.Context, msg kafka.Message) error {
								f.handlerRecv <- msg
								return fmt.Errorf("error")
							},
							Attempts: 1,
						},
					},
				}
			},
			msgs: []kafka.Message{
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
			},
			handlermsg: []kafka.Message{
				{
					Key:   []byte("key"),
					Value: []byte("value"),
				},
			},
			topicrmsg: []kafka.Message{},
			topicdmsg: []kafka.Message{},
			topicoff:  -1,
			topicroff: -1,
			topicdoff: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := fields{}
			if tt.prepare != nil {
				tt.prepare(&f)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			intr, err := interop.NewInterop([]string{broker}, f.flow, f.gc)
			require.NoError(t, err)
			done := make(chan struct{})
			go func() {
				require.Equal(t, tt.wantErr, intr.Start(ctx) != nil)
				done <- struct{}{}
			}()

			writer := kafka.Writer{
				Addr:  kafka.TCP(broker),
				Topic: f.topic,
			}
			for _, msg := range tt.msgs {
				require.NoError(t, writer.WriteMessages(ctx, msg))
			}

			g, ctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				for _, want := range tt.handlermsg {
					got := <-f.handlerRecv
					require.Equal(t, want.Key, got.Key)
					require.Equal(t, want.Value, got.Value)
				}
				return nil
			})
			g.Go(func() error {
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers: []string{broker},
					Topic:   f.topicr,
				})

				for _, want := range tt.topicrmsg {
					got, err := reader.ReadMessage(ctx)
					require.NoError(t, err)
					require.Equal(t, want.Key, got.Key)
					require.Equal(t, want.Value, got.Value)
				}
				return nil
			})
			g.Go(func() error {
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers: []string{broker},
					Topic:   f.topicd,
				})

				for _, want := range tt.topicdmsg {
					got, err := reader.ReadMessage(ctx)
					require.NoError(t, err)
					require.Equal(t, want.Key, got.Key)
					require.Equal(t, want.Value, got.Value)
				}
				return nil
			})
			require.NoError(t, g.Wait())
			cancel()
			<-done

			client := kafka.Client{
				Addr: kafka.TCP(broker),
			}
			resp, err := client.OffsetFetch(context.Background(), &kafka.OffsetFetchRequest{
				Addr:    kafka.TCP(broker),
				GroupID: f.gc,
				Topics: map[string][]int{
					f.topic:  {0},
					f.topicr: {0},
					f.topicd: {0},
				},
			})
			require.NoError(t, err)
			require.NoError(t, resp.Error)
			require.Equal(t, tt.topicoff, resp.Topics[f.topic][0].CommittedOffset)
			require.Equal(t, tt.topicroff, resp.Topics[f.topicr][0].CommittedOffset)
			require.Equal(t, tt.topicdoff, resp.Topics[f.topicd][0].CommittedOffset)
		})
	}
}
