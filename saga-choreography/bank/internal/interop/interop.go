package interop

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

type Rule struct {
	Handler  func(ctx context.Context, msg kafka.Message) error
	DLQ      string // if dlq is empty, returns error on failure.
	Attempts int    // retry attempts before sending to DLQ.
}

type Flow struct {
	Rules map[string]Rule
}

func listenTopics(flow Flow) []string {
	topics := make([]string, 0, len(flow.Rules))
	for topic := range flow.Rules {
		topics = append(topics, topic)
	}

	return topics
}

func NewInterop(brokers []string, flow Flow, gc string) (*Interop, error) {
	return &Interop{
		flow: flow,
		gc:   gc,
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupTopics: listenTopics(flow),
			GroupID:     gc,
		}),
	}, nil
}

type ireader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, messages ...kafka.Message) error
	Close() error
}

type iwriter interface {
	WriteMessages(ctx context.Context, messages ...kafka.Message) error
	Close() error
}

type Interop struct {
	flow   Flow
	gc     string
	reader ireader
	writer iwriter
}

func (i *Interop) Start(ctx context.Context) error {
	errc := make(chan error, 1)

	go func() {
		for {
			msg, err := i.reader.FetchMessage(ctx)
			if err == io.EOF {
				errc <- nil
			} else if err != nil {
				errc <- fmt.Errorf("failed fetch message: %w", err)
				return
			}

			rule, ok := i.flow.Rules[msg.Topic]
			if !ok {
				errc <- fmt.Errorf("no rule for topic: %s", msg.Topic)
				return
			}

			//// TODO(ezo): not sexy
			msg.Headers = setAttempts(msg.Headers, getAttempts(msg.Headers)+1)
			if err := rule.Handler(ctx, msg); err != nil {
				if err := i.retry(ctx, msg, err); err != nil {
					errc <- err
					return
				}
			}

			if err := i.reader.CommitMessages(ctx, msg); err != nil {
				errc <- err
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		if err := i.shutdown(); err != nil {
			return fmt.Errorf("failed shutdown: %w", err)
		}
	case err := <-errc:
		return err
	}

	return nil
}

const attemptsHeader = "attempts"

func getAttempts(headers []kafka.Header) int {
	for _, header := range headers {
		if header.Key == attemptsHeader {
			if val, err := strconv.Atoi(string(header.Value)); err != nil {
				log.Printf("failed to parse attempts header: %s", err)
			} else {
				return val
			}
		}
	}

	return 0
}

func setAttempts(headers []kafka.Header, num int) []kafka.Header {
	// To prevent change origin data
	nhs := append([]kafka.Header{}, headers...)
	for i, h := range nhs {
		if h.Key == attemptsHeader {
			nhs[i].Value = []byte(strconv.Itoa(num))
			return nhs
		}
	}

	return append(nhs, kafka.Header{
		Key:   attemptsHeader,
		Value: []byte(strconv.Itoa(num)),
	})
}

func (i *Interop) retry(ctx context.Context, msg kafka.Message, err error) error {
	attempts := getAttempts(msg.Headers)
	rule := i.flow.Rules[msg.Topic]

	if attempts >= rule.Attempts {
		if rule.DLQ == "" {
			return err
		}

		msg.Topic = rule.DLQ
		msg.Headers = setAttempts(msg.Headers, 0)
	}

	if err := i.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (i *Interop) shutdown() error {
	g := errgroup.Group{}
	g.Go(func() error {
		return i.reader.Close()
	})
	g.Go(func() error {
		return i.writer.Close()
	})

	return g.Wait()
}

func CreateTopic(brokers []string, topics ...string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker: %v", err)
	}

	// Hack to not get a "Not Available: the cluster is in the middle" error in WriteMessages.
	// Create or check if topics exist.
	wg := errgroup.Group{}
	for _, topic := range topics {
		topic := topic
		wg.Go(func() error {
			return conn.CreateTopics(
				kafka.TopicConfig{
					Topic:             topic,
					NumPartitions:     1,
					ReplicationFactor: 1,
				},
			)
		})
	}
	return wg.Wait()
}
