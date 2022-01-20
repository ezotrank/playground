package main

import (
	"database/sql"
	"embed"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

func load(db *sql.DB) error {
	sourceInstance, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}

func main() {
	c, err := pgx.ParseConfig("postgres://user:password@localhost:5432/db?sslmode=disable")
	if err != nil {
		log.Panicf("parsing postgres URI: %s", err)
	}

	db := stdlib.OpenDB(*c)

	if err := load(db); err != nil {
		panic(err)
	}
	log.Println("migration is done")
}
