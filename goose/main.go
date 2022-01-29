package main

import (
	"database/sql"
	"embed"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"

	_ "github.com/ezotrank/playground/goose/migrations"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
	_ = os.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/db?sslmode=disable")
}

func main() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Panicf("failed to open database: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
