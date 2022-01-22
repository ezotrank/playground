package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4/stdlib"
)

//go:embed migrations/*.sql
var fs embed.FS

func load(db *sql.DB) error {
	sourceInstance, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("failed to init iofs: %w", err)
	}

	driver, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return fmt.Errorf("failed to init postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to crate new migrate instance: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	if err := sourceInstance.Close(); err != nil {
		return fmt.Errorf("failed to close migration driver: %w", err)
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Panicf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := load(db); err != nil {
		log.Panicf("failed to load migrations: %v", err)
	}

	type user struct {
		ID        int
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	srv := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodGet:
				users := make([]user, 0)
				rows, err := db.Query("SELECT id, name, created_at, updated_at FROM users ORDER BY ID")
				if err != nil {
					log.Printf("failed to query users: %v", err)
					http.Error(w, "failed to query users", http.StatusInternalServerError)
					return
				}
				defer rows.Close()

				for rows.Next() {
					var u user
					err = rows.Scan(&u.ID, &u.Name, &u.CreatedAt, &u.UpdatedAt)
					if err != nil {
						panic(err)
					}
					users = append(users, u)
				}

				if rows.Err() != nil {
					panic(err)
				}

				if err := json.NewEncoder(w).Encode(users); err != nil {
					log.Printf("failed to encode users: %v", err)
					http.Error(w, "failed to encode users", http.StatusInternalServerError)
					return
				}
			case http.MethodPost:
				var u user
				if err := json.NewDecoder(req.Body).Decode(&u); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				_, err := db.Exec("INSERT INTO users (name) VALUES ($1)", u.Name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fmt.Fprintf(w, "User %s created", u.Name)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}),
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
