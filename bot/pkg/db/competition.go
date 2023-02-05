package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Competition struct {
	Id        string    `db:"id"`
	ChatID    int64     `db:"chat_id"`
	Users     []string  `db:"users"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *Repo) RegisterCompetition(ctx context.Context, competition Competition) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.RegisterCompetitionTxx(tx, competition)
}

func (r *Repo) RegisterCompetitionTxx(tx *sqlx.Tx, competition Competition) error {
	competition.CreatedAt = time.Now()
	competition.UpdatedAt = time.Now()
	stmt, err := tx.Prepare(`INSERT INTO competition
		(chat_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare competition stmt")
	}
	defer stmt.Close()
	var id string
	err = stmt.QueryRow(
		competition.ChatID,
		competition.StartDate,
		competition.EndDate,
		competition.CreatedAt,
		competition.UpdatedAt,
	).Scan(&id)
	for _, user := range competition.Users {
		queryUsers := `INSERT INTO competition_users
		(user_id, competition_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`
		_, err := tx.Exec(queryUsers, user, id, competition.CreatedAt, competition.UpdatedAt)
		if err != nil {
			return errors.Wrapf(err, "failed to insert user %s", user)
		}
	}
	return nil
}

func (r *Repo) GetCompetitionByChat(ctx context.Context, chatID int64) (*Competition, error) {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.GetCompetitionByChatTxx(tx, chatID)
}

func (r *Repo) GetCompetitionByChatTxx(tx *sqlx.Tx, chatID int64) (*Competition, error) {
	query := `SELECT 
		id, chat_id, start_date, end_date, created_at, updated_at 
		FROM competition WHERE chat_id = $1 LIMIT 1;`
	record := &Competition{}
	err := tx.Get(record, query, chatID)
	return record, err
}

func (r *Repo) RegisterUserInCompetition(ctx context.Context, userID string, competitionID string) error {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return errors.Wrap(err, "failed to begin db transaction")
	}
	defer tx.Commit() //nolint:errcheck
	return r.RegisterUserInCompetitionTxx(tx, userID, competitionID)
}

func (r *Repo) RegisterUserInCompetitionTxx(tx *sqlx.Tx, userID string, competitionID string) error {
	query := `INSERT INTO competition_users
		(user_id, competition_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(query, userID, competitionID, time.Now(), time.Now())
	if err != nil {
		return errors.Wrapf(err, "failed to register user '%s' in competition '%s'", userID, competitionID)
	}
	return nil
}
