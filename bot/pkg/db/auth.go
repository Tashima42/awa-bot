package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Auth struct {
	Id        string    `db:"id"`
	UserID    string    `db:"user_id"`
	ApiKey    string    `db:"api_key"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *Repo) RegisterApiKey(ctx context.Context, auth Auth) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.RegisterApiKeyTxx(tx, auth)
}

func (r *Repo) RegisterApiKeyTxx(tx *sqlx.Tx, auth Auth) error {
	query := `INSERT INTO auth
		(user_id, api_key)
		VALUES ($1, $2)`
	_, err := tx.Exec(query, auth.UserID, auth.ApiKey)
	if err != nil {
		return errors.Wrapf(err, "failed to insert auth for %s", auth.UserID)
	}
	return nil
}

func (r *Repo) DeleteApiKey(ctx context.Context, userID string) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.DeleteApiKeyTxx(tx, userID)
}

func (r *Repo) DeleteApiKeyTxx(tx *sqlx.Tx, userID string) error {
	query := `DELETE FROM auth WHERE user_id = $1`
	_, err := tx.Exec(query, userID)
	if err != nil {
		return errors.Wrapf(err, "failed to remove auth for %s", userID)
	}
	return nil
}

func (r *Repo) GetApiKeyByUserIdTxx(tx *sqlx.Tx, userID string) (string, error) {
	var apiKey string
	query := `SELECT api_key FROM auth WHERE user_id = $1 LIMIT 1;`
	err := tx.Get(&apiKey, query, userID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get api key for %s", userID)
	}
	return apiKey, nil
}
