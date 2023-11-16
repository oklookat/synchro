package repository

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/streaming"
)

var (
	//go:embed library.sql
	librarySQL string

	_db  *sqlx.DB
	_log *logger.Logger

	Services map[streaming.ServiceName]streaming.Service
)

func Boot(dbPath string, log *logger.Logger, services map[streaming.ServiceName]streaming.Service) error {
	_log = log

	if _, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			dbFile, err := os.Create(dbPath)
			if err != nil {
				return err
			}
			dbFile.Close()
		} else {
			return err
		}
	}

	var err error

	if _db, err = sqlx.Open("sqlite3", dbPath); err != nil {
		return err
	}
	if _, err = dbExec(context.Background(), librarySQL); err != nil {
		return err
	}

	Services = make(map[streaming.ServiceName]streaming.Service, len(services))
	for name := range services {
		// Create / get.
		repo, err := newOrExistingServiceDatabase(services[name])
		if err != nil {
			return err
		}

		// Boot.
		if err := services[name].Boot(repo); err != nil {
			return fmt.Errorf("boot '%s': %w", name, err)
		}

		Services[name] = services[name]
	}

	return err
}

func DebugCleanAllExceptAccounts() error {
	_, err := dbExec(context.Background(), `
	DROP TABLE artist;
	DROP TABLE linked_artist;

	DROP TABLE album;
	DROP TABLE linked_album;

	DROP TABLE track;
	DROP TABLE linked_track;
	`)
	return err
}
