package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4/source"
	"log"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Repo struct {
	db      *sqlx.DB
	config  Config
	migrate source.Driver
}

// Config database connection configuration
type Config struct {
	Port     int
	Host     string
	DBName   string
	User     string
	Password string
}

func (c *Config) FromEnv() {
	log.Println("getting config from env")
	if os.Getenv("DB_HOST") != "" {
		log.Println("found DB_HOST")
		c.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_NAME") != "" {
		log.Println("found DB_NAME")
		c.DBName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_PASSWORD") != "" {
		log.Println("found DB_PASSWORD")
		c.Password = os.Getenv("DB_PASSWORD")
	}
	if os.Getenv("DB_USER") != "" {
		log.Println("found DB_USER")
		c.User = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PORT") != "" {
		log.Println("found DB_PORT")
		port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
		c.Port = port
	}
	if c.Port == 0 {
		log.Println("port = 0, using default postgres port 5432")
		c.Port = 5432
	}
}

func Open(conf Config) (*Repo, error) {
	log.Println("connecting to database")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Password, conf.DBName)
	var err error
	repo := &Repo{}

	log.Println("openning postgres")
	if repo.db, err = sqlx.Open("postgres", psqlInfo); err != nil {
		return nil, errors.Wrap(err, "could not open postgres")
	}

	log.Println("testing db connection")
	if err := repo.db.Ping(); err != nil {
		return nil, errors.Wrapf(err, "could not ping db: host=%s db=%s user=%s ", conf.Host, conf.DBName, conf.User)
	}
	repo.config = conf
	return repo, nil
}

// Close closes the connections for the tenant
func (r *Repo) Close() error {
	log.Println("closing db")
	return r.db.Close()
}

// BeginTxx returns new admin db transaction
func (r *Repo) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	if opts == nil {
		opts = &sql.TxOptions{}
	}

	return r.db.BeginTxx(ctx, opts)
}

func Rollback(tx *sqlx.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = errors.Wrap(err, rerr.Error())
	}
	return err
}
