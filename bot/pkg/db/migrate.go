package db

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/hashicorp/go-multierror"
	"log"
)

func (r *Repo) Up() (version uint, dirty bool, err error) {
	log.Println("getting migrate")
	m, err := r.migration()
	if err != nil {
		return 0, false, err
	}
	defer func() {
		_, _ = m.Close()
	}()

	log.Println("migrating up")
	mErr := m.Up()
	if mErr == migrate.ErrNoChange {
		mErr = nil
	}
	log.Println("getting version")
	resultVersion, resultDirty, resultErr := m.Version()
	return resultVersion, resultDirty, multierror.Append(mErr, resultErr).ErrorOrNil()
}

func (r *Repo) Version(version uint) (currentVersion uint, dirty bool, err error) {
	m, err := r.migration()
	if err != nil {
		return 0, false, err
	}
	defer func() {
		_, _ = m.Close()
	}()
	mErr := m.Migrate(version)
	if mErr == migrate.ErrNoChange {
		mErr = nil
	}
	resultVersion, resultDirty, resultErr := m.Version()
	return resultVersion, resultDirty, multierror.Append(mErr, resultErr).ErrorOrNil()
}

func (r *Repo) migration() (*migrate.Migrate, error) {
	return migrate.NewWithSourceInstance("httpfs", r.migrate, r.databaseURL())
}

func (r *Repo) databaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		r.config.User, r.config.Password, r.config.Host, r.config.Port, r.config.DBName,
	)
}
