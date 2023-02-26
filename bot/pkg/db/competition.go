package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"log"
	"time"
)

type Competition struct {
	Id        string         `db:"id"`
	ChatID    int64          `db:"chat_id"`
	Users     pq.StringArray `db:"users"`
	StartDate time.Time      `db:"start_date"`
	EndDate   time.Time      `db:"end_date"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func (c *Competition) IsUserRegistered(userId string) bool {
	for _, u := range c.Users {
		if u == userId {
			return true
		}
	}
	return false
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
	if err != nil {
		return err
	}
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
			c.id AS id,
			c.chat_id AS chat_id,
			c.start_date AS start_date,
			c.end_date AS end_date,
			c.created_at AS created_at,
			c.updated_at AS updated_at,
			array_agg(cu.user_id) AS users
		FROM competition c
		LEFT JOIN competition_users cu ON cu.competition_id = c.id
		WHERE chat_id = $1
		GROUP BY c.id`
	record := &Competition{}
	err := tx.Get(record, query, chatID)
	return record, err
}

func (r *Repo) RegisterUsersInCompetitionTxx(tx *sqlx.Tx, users []string, competitionID string) error {
	query := "INSERT INTO competition_users (user_id, competition_id, created_at, updated_at) VALUES"
	var values []interface{}
	now := time.Now()
	for _, user := range users {
		values = append(values, user, competitionID, now, now)
	}
	numberOfFields := 4
	for i := range values {
		if i%numberOfFields == 0 {
			query += fmt.Sprintf("($%d, $%d, $%d, $%d)", i+1, i+2, i+3, i+4)
			if i != len(values)-1 && i != 0 && i%numberOfFields == 0 {
				query += ","
			}
		}
	}
	log.Println(query, values)
	_, err := tx.Exec(query, values...)
	if err != nil {
		return errors.Wrapf(err, "failed to register users in competition '%s'", competitionID)
	}
	return nil
}
