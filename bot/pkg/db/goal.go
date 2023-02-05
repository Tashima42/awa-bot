package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Goal struct {
	Id        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Goal      int       `db:"goal"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *Repo) RegisterGoal(ctx context.Context, goal Goal) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.RegisterGoalTxx(tx, goal)
}

func (r *Repo) RegisterGoalTxx(tx *sqlx.Tx, goal Goal) error {
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()
	query := `INSERT INTO goals
		(user_id, goal, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(query, goal.UserID, goal.Goal, goal.CreatedAt, goal.UpdatedAt)
	if err != nil {
		return errors.Wrapf(err, "failed to insert goal for %d", goal.UserID)
	}
	return nil
}

func (r *Repo) GetGoalByUser(ctx context.Context, userID string) (*Goal, error) {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.GetGoalByUserTxx(tx, userID)
}

func (r *Repo) GetGoalByUserTxx(tx *sqlx.Tx, userID string) (*Goal, error) {
	query := `SELECT id, user_id, goal, created_at, updated_at 
		FROM goals WHERE user_id = $1 
	   	ORDER BY updated_at desc LIMIT 1;`
	record := &Goal{}
	err := tx.Get(record, query, userID)
	return record, err
}
