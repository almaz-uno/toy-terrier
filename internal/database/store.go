package database

import (
	"context"
	"database/sql"

	db "toy-terrier-telegram/internal/database/sqlc"
)

// Store wraps sqlc queries with transaction support
type Store struct {
	*db.Queries
	db *sql.DB
}

// NewStore creates a new store instance
func NewStore(database *sql.DB) *Store {
	return &Store{
		Queries: db.New(database),
		db:      database,
	}
}

// ExecTx executes a function within a database transaction
func (store *Store) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := db.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
