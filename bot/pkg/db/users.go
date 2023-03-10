package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type User struct {
	Id         string    `db:"id" json:"id"`
	TelegramID int64     `db:"telegram_id" json:"telegram_id"`
	Name       string    `db:"name" json:"name"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

func (r *Repo) RegisterUser(ctx context.Context, user User) error {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.RegisterUserTxx(tx, user)
}

func (r *Repo) RegisterUserTxx(tx *sqlx.Tx, user User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	query := `INSERT INTO users 
		(telegram_id, created_at, updated_at) 
		VALUES (:telegram_id, :created_at, :updated_at);`
	_, err := tx.NamedExec(query, user)
	return err
}

func (r *Repo) GetUser(ctx context.Context, telegramID int64) (*User, error) {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.GetUserTxx(tx, telegramID)
}

func (r *Repo) GetUserTxx(tx *sqlx.Tx, telegramID int64) (*User, error) {
	query := `SELECT 
		id AS id, 
		telegram_id AS telegram_id, 
		created_at AS created_at, 
		updated_at AS updated_at 
	FROM users 
	WHERE telegram_id = $1 
	LIMIT 1;`
	record := &User{}
	err := tx.Get(record, query, telegramID)
	return record, err
}

func (r *Repo) GetUserByID(ctx context.Context, userID string) (*User, error) {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.GetUserByIDTxx(tx, userID)
}

func (r *Repo) GetUserByIDTxx(tx *sqlx.Tx, userID string) (*User, error) {
	query := `SELECT 
		id AS id, 
		telegram_id AS telegram_id, 
		created_at AS created_at, 
		updated_at AS updated_at 
	FROM users 
	WHERE id = $1 
	LIMIT 1;`
	record := &User{}
	err := tx.Get(record, query, userID)
	return record, err
}
