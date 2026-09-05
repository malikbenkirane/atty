package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateTableHistory, downCreateTableHistory)
}

func upCreateTableHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
			CREATE TABLE access_history (
				id INTEGER PRIMARY KEY,
				title TEXT,
				accessed_at INTEGER
			)
		`)
	if err != nil {
		return fmt.Errorf("tx: exec: create access_history table: %w", err)
	}
	return nil
}

func downCreateTableHistory(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
