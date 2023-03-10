package db

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Water struct {
	Id        string    `db:"id" json:"id"`
	UserId    string    `db:"user_id" json:"user_id"`
	Amount    int       `db:"amount" json:"amount"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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

func (r *Repo) GetUserWaterPaginated(ctx context.Context, userID string, limit int, skip int) (water []*Water, total int, err error) {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Commit() //nolint:errcheck
	return r.GetUserWaterPaginatedTxx(tx, userID, limit, skip)
}

func (r *Repo) GetUserWaterPaginatedTxx(tx *sqlx.Tx, userID string, limit int, skip int) (water []*Water, total int, err error) {
	var waters []*Water
	var count int
	query := `SELECT 
					id, 	
					user_id, 
					amount, 
					created_at, 
					updated_at 
				FROM water 
				WHERE user_id = $1 
				ORDER BY created_at DESC 
				LIMIT $2 
				OFFSET $3`
	countQuery := "SELECT COUNT(*) FROM water WHERE user_id = $1"
	err = tx.Select(&waters, query, userID, limit, skip)
	if err != nil {
		return nil, 0, err
	}
	err = tx.Get(&count, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	return waters, count, nil
}

func todayAtMidnight() string {
	formatedDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s 00:00:00.000000", formatedDate)
}
