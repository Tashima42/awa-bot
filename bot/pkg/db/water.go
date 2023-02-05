package db

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Water struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Amount    int       `db:"amount"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type TimePast string

const (
	Today       TimePast = "today"
	StartOfWeek TimePast = "startofweek"
)

func (r *Repo) RegisterWater(ctx context.Context, water Water) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.RegisterWaterTxx(tx, water)
}
func (r *Repo) RegisterWaterTxx(tx *sqlx.Tx, water Water) error {
	water.CreatedAt = time.Now()
	water.UpdatedAt = time.Now()
	query := `INSERT INTO water 
		(user_id, amount, created_at, updated_at)
		VALUES (:user_id, :amount, :created_at, :updated_at);`
	_, err := tx.NamedExec(query, water)
	return err
}
func (r *Repo) GetUserAmountTxx(tx *sqlx.Tx, userID string, timePast TimePast) (*int64, error) {
	var amount int64
	timeArg := todayAtMidnight()
	query := `SELECT coalesce(sum(amount), 0) AS sum FROM water WHERE user_id = $1 AND created_at >= $2`
	err := tx.Get(&amount, query, userID, timeArg)
	if err != nil {
		return nil, err
	}
	fmt.Println(amount)
	return &amount, nil
}

func todayAtMidnight() string {
	formatedDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s 00:00:00.000000", formatedDate)
}
