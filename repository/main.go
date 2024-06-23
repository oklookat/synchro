package repository

import (
	"context"
	_ "embed"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklookat/synchro/shared"
)

const (
	EntityNameArtist   EntityName = "artist"
	EntityNameAlbum    EntityName = "album"
	EntityNameTrack    EntityName = "track"
	EntityNamePlaylist EntityName = "playlist"
)

var (
	Remotes map[shared.RemoteName]shared.Remote

	ArtistEntity   = NewEntityRepository(EntityNameArtist)
	AlbumEntity    = NewEntityRepository(EntityNameAlbum)
	TrackEntity    = NewEntityRepository(EntityNameTrack)
	PlaylistEntity = NewEntityRepository(EntityNamePlaylist)

	ArtistLinkable = NewEntityRepository(EntityNameArtist)
	AlbumLinkable  = NewEntityRepository(EntityNameAlbum)
	TrackLinkable  = NewEntityRepository(EntityNameTrack)

	ArtistSyncable   = NewSyncableEntity(EntityNameArtist)
	AlbumSyncable    = NewSyncableEntity(EntityNameAlbum)
	TrackSyncable    = NewSyncableEntity(EntityNameTrack)
	PlaylistSyncable = NewSyncableEntity(EntityNamePlaylist)
)

func NewLinkablePlaylist(accountID uint64) *LinkableEntity {
	return NewLinkableEntity("playlist", shared.RemoteName(strconv.FormatUint(accountID, 10)))
}

const (
	_packageName = "repository"
)

var (
	//go:embed library.sql
	_librarySQL string
	_db         *sqlx.DB
)

func Boot(remotes map[shared.RemoteName]shared.Remote) error {
	const dbPath = "data.sqlite"

	_, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	if _db, err = sqlx.Open("sqlite3", dbPath); err != nil {
		return err
	}

	if _, err = dbExec(context.Background(), _librarySQL); err != nil {
		return err
	}

	Remotes = make(map[shared.RemoteName]shared.Remote, len(remotes))

	for name := range remotes {
		// Create / get.
		repo, err := newOrExistingRemote(remotes[name])
		if err != nil {
			return err
		}

		// Boot.
		if err := remotes[name].Boot(repo); err != nil {
			continue
		}

		Remotes[name] = remotes[name]
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
