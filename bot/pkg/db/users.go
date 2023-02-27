package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type User struct {
	Id         string    `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Name       string    `db:"name"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
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
		name AS name,
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

func (r *Repo) GetUserByApiKey(ctx context.Context, apiKey string) (*User, error) {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.GetUserByApiKeyTxx(tx, apiKey)
}

func (r *Repo) GetUserByApiKeyTxx(tx *sqlx.Tx, apiKey string) (*User, error) {
	query := `SELECT 
		u.id AS id, 
		u.telegram_id AS telegram_id, 
		u.created_at AS created_at, 
		u.updated_at AS updated_at 
	FROM auth a
	JOIN users u on a.user_id = u.id
	WHERE a.api_key = $1 
	LIMIT 1;`
	record := &User{}
	err := tx.Get(record, query, apiKey)
	return record, err
}
