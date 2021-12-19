/*

 */

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dsn           = "postgres://postgres@localhost:31721/db?pool_max_conns=100"
	addr          = ":23123"
	eventsChannel = "events_status_channel"
	consumers     = 10
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalln("failed to parse config:", err)
	}
	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	srv := http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			switch r.URL.Path {
			case "/set":
				const query = `INSERT INTO events (key, status, created_at, updated_at) VALUES ($1, 'new', now(), now())`
				if _, err := pool.Exec(r.Context(), query, key); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "/get":
				const query = `SELECT status FROM events WHERE key = $1`
				var status string
				err = pool.QueryRow(r.Context(), query, key).Scan(&status)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, _ = fmt.Fprintf(w, "event status: %s", status)
			default:
				http.Error(w, "not found", http.StatusNotFound)
			}
		}),
	}

	go func() {
		for i := 0; i < consumers; i++ {
			go consumer(ctx, pool, i)
			log.Println("start consumer", i+1)
		}
	}()

	go func() {
		log.Println("starting listening server on", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("failed to start http server: %v", err)
		}
	}()

	<-make(chan struct{})
}

func process(_ int) {
	time.Sleep(time.Millisecond)
}

func consumer(ctx context.Context, pool *pgxpool.Pool, id int) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "LISTEN "+eventsChannel)
	if err != nil {
		log.Fatalf("failed to listen channel: %v", err)
	}

	for {
		msg, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			log.Fatalf("failed to wait for msg: %v", err)
		}

		log.Printf("consumer %d was notified", id)

		// Payload contains only status, and this consumer works with "new" status only.
		if msg.Payload != "new" {
			continue
		}

		const query = `
			UPDATE events SET status='running', updated_at=now()
			WHERE id = (
			  SELECT id
			  FROM events
			  WHERE status='new'
			  ORDER BY id
			  FOR UPDATE SKIP LOCKED
			  LIMIT 1
			)
			RETURNING id;
			`

		var evid int
		err = conn.QueryRow(context.Background(), query).Scan(&evid)
		if err != nil && err == pgx.ErrNoRows {
			continue
		} else if err != nil {
			log.Fatalln("failed to update event status to running:", err)
		}

		// Some random time to simulate work.
		process(evid)

		if num := rand.Intn(100); num == 0 {
			_, err = conn.Exec(context.Background(), "UPDATE events SET status='error', error=$1, updated_at=now() WHERE id = $2", "error happens", evid)
			if err != nil {
				log.Fatalln("failed to update event status:", err)
			}
		} else {
			_, err = conn.Exec(context.Background(), "UPDATE events SET status='success', updated_at=now() WHERE id = $1", evid)
			if err != nil {
				log.Fatalln("failed to update event status:", err)
			}
		}

		log.Printf("consumder %d processed message", id)
	}
}
