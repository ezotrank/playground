package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
	"github.com/ory/dockertest/v3"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/ezotrank/playground/saga-choreography/wallet/gen"
)

var resources map[string]*dockertest.Resource

func TestMain(m *testing.M) {
	res, down := GetPool()
	resources = res
	code := m.Run()
	down()
	os.Exit(code)
}

func purge(t *testing.T) {
	if err := FlushRedis(resources["redis"]); err != nil {
		t.Fatalf("failed flush redis %s", err)
	}
}

func getService(ctx context.Context, srv *Server) (pb.WalletServiceClient, error) {
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(srv)))
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	return pb.NewWalletServiceClient(conn), nil
}

func dialer(srv *Server) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterWalletServiceServer(server, srv)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func GetPool() (map[string]*dockertest.Resource, func()) {
	if err := setupNetwork(network); err != nil {
		log.Fatalf("failed to setup network: %v", err)
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("failed connect to docker: %s", err)
	}

	resources := map[string]*dockertest.Resource{
		"redis":     RedisStart(pool),
		"zookeeper": ZooKeeperStart(pool),
		"kafka":     KafkaStart(pool),
	}

	cancel := func() {
		for _, r := range resources {
			if err := r.Close(); err != nil {
				log.Printf("failed close resource: %s", err)
			}
		}
	}

	return resources, cancel
}

func setupNetwork(network string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts()
	if err != nil {
		return err
	}
	defer cli.Close()

	resp, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for i := range resp {
		if resp[i].Name == network {
			return nil
		}
	}

	_, err = cli.NetworkCreate(ctx, network, types.NetworkCreate{CheckDuplicate: true})
	return err
}

func RedisStart(pool *dockertest.Pool) *dockertest.Resource {
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       "redis",
		Repository: "redis",
		Tag:        "5-alpine",
		NetworkID:  network,
		Hostname:   "redis",
	})
	if err != nil {
		log.Fatalf("failed start redis: %s", err)
	}

	if err = pool.Retry(func() error {
		client := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", res.GetPort("6379/tcp")),
		})

		return client.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("failed connect to redis: %s", err)
	}

	return res
}

func FlushRedis(res *dockertest.Resource) error {
	status := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%s", res.GetPort("6379/tcp")),
	}).FlushAll(context.Background())
	if err := status.Err(); err != nil {
		return fmt.Errorf("faild flush redis: %w", err)
	}

	return nil
}

func ZooKeeperStart(pool *dockertest.Pool) *dockertest.Resource {
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       "zookeeper",
		Repository: "wurstmeister/zookeeper",
		Tag:        "latest",
		NetworkID:  network,
		Hostname:   "zookeeper",
	})
	if err != nil {
		log.Fatalf("failed start redis: %s", err)
	}

	conn, _, err := zk.Connect([]string{fmt.Sprintf("127.0.0.1:%s", res.GetPort("2181/tcp"))}, 10*time.Second)
	if err != nil {
		log.Fatalf("could not connect zookeeper: %s", err)
	}
	defer conn.Close()

	retryFn := func() error {
		switch conn.State() {
		case zk.StateHasSession, zk.StateConnected:
			return nil
		default:
			return errors.New("not yet connected")
		}
	}

	if err = pool.Retry(retryFn); err != nil {
		log.Fatalf("could not connect to zookeeper: %s", err)
	}

	return res
}

const network = "wallet-test"

func KafkaStart(pool *dockertest.Pool) *dockertest.Resource {
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:         "kafka",
		Repository:   "wurstmeister/kafka",
		Tag:          "2.13-2.8.1",
		ExposedPorts: []string{"9094/tcp"},
		NetworkID:    network,
		Hostname:     "kafka",
		Env: []string{
			"KAFKA_CREATE_TOPICS=health-check:1:1:compact",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT",
			"KAFKA_LISTENERS=INSIDE://0.0.0.0:9092,OUTSIDE://0.0.0.0:9094",
			"KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181",
			"KAFKA_INTER_BROKER_LISTENER_NAME=INSIDE",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
			"PORT_COMMAND=docker port $(hostname) 9094/tcp | cut -d: -f2",
			"KAFKA_ADVERTISED_LISTENERS=INSIDE://kafka:9092,OUTSIDE://localhost:_{PORT_COMMAND}",
		},
		Mounts: []string{"/var/run/docker.sock:/var/run/docker.sock"},
	})
	if err != nil {
		log.Fatalf("could not start kafka: %s", err)
	}

	retryFn := func() error {
		writer := &kafka.Writer{
			Addr:  kafka.TCP(fmt.Sprintf("localhost:%s", res.GetPort("9094/tcp"))),
			Topic: "health-check",
		}
		defer writer.Close()
		err := writer.WriteMessages(context.Background(), kafka.Message{Value: []byte("ping")})
		if err != nil {
			return err
		}

		return nil
	}

	if err = pool.Retry(retryFn); err != nil {
		log.Fatalf("could not connect to kafka: %s", err)
	}

	return res
}
