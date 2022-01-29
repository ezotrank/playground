package migrations

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddUsers, downAddUsers)
}

func upAddUsers(tx *sql.Tx) error {
	for i := 0; i < 1000; i++ {
		if _, err := tx.Exec(
			`INSERT INTO "user"(name, email) VALUES ($1, $2)`,
			fmt.Sprintf("user%d", i),
			fmt.Sprintf("user%d@example.com", i),
		); err != nil {
			return err
		}
	}
	return nil
}

func downAddUsers(tx *sql.Tx) error {
	if _, err := tx.Exec(`DELETE FROM "user" WHERE name LIKE 'user%'`); err != nil {
		return err
	}
	return nil
}
