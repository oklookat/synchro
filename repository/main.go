package repository

import (
	"context"
	_ "embed"
	"errors"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklookat/synchro/streaming"
	"github.com/rs/zerolog"
)

var (
	//go:embed library.sql
	librarySQL string

	_db  *sqlx.DB
	_log zerolog.Logger

	Remotes map[streaming.ServiceName]streaming.Service
)

const (
	_packageName = "repository"
)

func Boot(dbPath string, log zerolog.Logger, services map[streaming.ServiceName]streaming.Service) error {
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

	Remotes = make(map[streaming.ServiceName]streaming.Service, len(services))
	for name := range services {
		// Create / get.
		repo, err := newOrExistingRemote(services[name])
		if err != nil {
			return err
		}

		// Boot.
		if err := services[name].Boot(repo); err != nil {
			return err
		}

		Remotes[name] = services[name]
	}

	return err
}

func DebugCleanAllExceptAccounts() error {
	_, err := dbExec(context.Background(), `
	DROP TABLE snapshot;
	DROP TABLE snapshot_liked_album;
	DROP TABLE snapshot_liked_artist;
	DROP TABLE snapshot_liked_track;
	DROP TABLE snapshot_playlist;
	DROP TABLE snapshot_playlist_track;

	DROP TABLE artist;
	DROP TABLE linked_artist;
	DROP TABLE synced_artist;

	DROP TABLE album;
	DROP TABLE linked_album;
	DROP TABLE synced_album;

	DROP TABLE playlist;
	DROP TABLE linked_playlist;
	DROP TABLE synced_playlist;
	DROP TABLE synced_playlist_track;
	DROP TABLE account_synced_playlist_settings;

	DROP TABLE track;
	DROP TABLE linked_track;
	DROP TABLE synced_track;
	`)
	return err
}
