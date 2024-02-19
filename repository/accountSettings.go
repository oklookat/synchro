package repository

import (
	"context"
	"time"

	"github.com/oklookat/synchro/shared"
)

func newAccountSettings(accountID shared.RepositoryID) (*AccountSettings, error) {
	setts := &AccountSettings{
		HAccountID: accountID,
	}
	err := setts.load()
	return setts, err
}

type AccountSettings struct {
	HID        shared.RepositoryID `db:"id"`
	HAccountID shared.RepositoryID `db:"account_id"`

	HSyncLikedAlbums  bool `db:"sync_liked_albums"`
	HSyncLikedArtists bool `db:"sync_liked_artists"`
	HSyncLikedTracks  bool `db:"sync_liked_tracks"`
	HSyncPlaylists    bool `db:"sync_playlists"`

	HLastSyncLikedAlbums  int64 `db:"last_sync_liked_albums"`
	HLastSyncLikedArtists int64 `db:"last_sync_liked_artists"`
	HLastSyncLikedTracks  int64 `db:"last_sync_liked_tracks"`
	HLastSyncPlaylists    int64 `db:"last_sync_playlists"`
}

func (e *AccountSettings) load() error {
	const (
		selectQuery = "SELECT * FROM account_settings WHERE account_id = ? LIMIT 1"
		insertQuery = "INSERT INTO account_settings (id, account_id) VALUES (?, ?) RETURNING *"
	)

	setts, err := dbGetOne[AccountSettings](context.Background(), selectQuery, genRepositoryID(), e.HAccountID)
	if err != nil {
		return err
	}

	if shared.IsNil(setts) {
		setts, err = dbGetOne[AccountSettings](context.Background(), insertQuery, e.HAccountID)
		if err != nil {
			return err
		}
	}

	*e = *setts
	return err
}

func (e *AccountSettings) Update(from *AccountSettings) error {
	if from == nil {
		return nil
	}
	if err := e.LikedAlbums().SetSynchronize(from.LikedAlbums().Synchronize()); err != nil {
		return err
	}
	if err := e.LikedArtists().SetSynchronize(from.LikedArtists().Synchronize()); err != nil {
		return err
	}
	if err := e.LikedTracks().SetSynchronize(from.LikedTracks().Synchronize()); err != nil {
		return err
	}
	if err := e.Playlists().SetSynchronize(from.Playlists().Synchronize()); err != nil {
		return err
	}
	return nil
}

func (e *AccountSettings) LikedAlbums() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_settings SET sync_liked_albums = ? WHERE account_id = ?"
		setLastSyncQuery = "UPDATE account_settings SET last_sync_liked_albums = ? WHERE account_id = ?"
	)
	return newBoolSyncSetting(e.HSyncLikedAlbums, func(val bool) error {
		e.HSyncLikedAlbums = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID)
		return err

	}, e.HLastSyncLikedAlbums, func(val int64) error {
		e.HLastSyncLikedAlbums = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID)
		return err
	})
}

func (e *AccountSettings) LikedArtists() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_settings SET sync_liked_artists = ? WHERE account_id = ?"
		setLastSyncQuery = "UPDATE account_settings SET last_sync_liked_artists = ? WHERE account_id = ?"
	)
	return newBoolSyncSetting(e.HSyncLikedArtists, func(val bool) error {
		e.HSyncLikedArtists = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID)
		return err
	}, e.HLastSyncLikedArtists, func(val int64) error {
		e.HLastSyncLikedArtists = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID)
		return err
	})
}

func (e *AccountSettings) LikedTracks() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_settings SET sync_liked_tracks = ? WHERE account_id = ?"
		setLastSyncQuery = "UPDATE account_settings SET last_sync_liked_tracks = ? WHERE account_id = ?"
	)
	return newBoolSyncSetting(e.HSyncLikedTracks, func(val bool) error {
		e.HSyncLikedTracks = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID)
		return err
	}, e.HLastSyncLikedTracks, func(val int64) error {
		e.HLastSyncLikedTracks = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID)
		return err
	})
}

func (e *AccountSettings) Playlists() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_settings SET sync_playlists = ? WHERE account_id = ?"
		setLastSyncQuery = "UPDATE account_settings SET last_sync_playlists = ? WHERE account_id = ?"
	)
	return newBoolSyncSetting(e.HSyncPlaylists, func(val bool) error {
		e.HSyncPlaylists = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID)
		return err
	}, e.HLastSyncPlaylists, func(val int64) error {
		e.HLastSyncPlaylists = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID)
		return err
	})
}

func (e *AccountSettings) Playlist(playlistID shared.EntityID) (shared.PlaylistSyncSettings, error) {
	return newAccountSyncedPlaylistSettings(playlistID, e.HAccountID)
}

func newAccountSyncedPlaylistSettings(playlistID shared.EntityID, accountID shared.RepositoryID) (*AccountSyncedPlaylistSettings, error) {
	setts := &AccountSyncedPlaylistSettings{
		HPlaylistID: playlistID,
		HAccountID:  accountID,
	}
	err := setts.load()
	return setts, err
}

type AccountSyncedPlaylistSettings struct {
	HID         shared.RepositoryID `db:"id"`
	HPlaylistID shared.EntityID     `db:"playlist_id"`
	HAccountID  shared.RepositoryID `db:"account_id"`

	HSyncName        bool `db:"sync_name"`
	HSyncDescription bool `db:"sync_description"`
	HSyncVisibility  bool `db:"sync_visibility"`
	HSyncTracks      bool `db:"sync_tracks"`

	HLastSyncName        int64 `db:"last_sync_name"`
	HLastSyncDescription int64 `db:"last_sync_description"`
	HLastSyncVisibility  int64 `db:"last_sync_visibility"`
	HLastSyncTracks      int64 `db:"last_sync_tracks"`
}

func (e *AccountSyncedPlaylistSettings) load() error {
	const (
		selectQuery = "SELECT * FROM account_synced_playlist_settings WHERE account_id = ? AND playlist_id = ? LIMIT 1"
		insertQuery = "INSERT INTO account_synced_playlist_settings (account_id, playlist_id) VALUES (?, ?) RETURNING *"
	)

	setts, err := dbGetOne[AccountSyncedPlaylistSettings](context.Background(), selectQuery, e.HAccountID, e.HPlaylistID)
	if err != nil {
		return err
	}

	if shared.IsNil(setts) {
		setts, err = dbGetOne[AccountSyncedPlaylistSettings](context.Background(), insertQuery, e.HAccountID, e.HPlaylistID)
		if err != nil {
			return err
		}
	}

	*e = *setts
	return err
}

func (e *AccountSyncedPlaylistSettings) Name() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_synced_playlist_settings SET sync_name = ? WHERE account_id = ? AND playlist_id = ?"
		setLastSyncQuery = "UPDATE account_synced_playlist_settings SET last_sync_name = ? WHERE account_id = ? AND playlist_id = ?"
	)
	return newBoolSyncSetting(e.HSyncName, func(val bool) error {
		e.HSyncName = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	}, e.HLastSyncName, func(val int64) error {
		e.HLastSyncName = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	})
}

func (e *AccountSyncedPlaylistSettings) Description() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_synced_playlist_settings SET sync_description = ? WHERE account_id = ? AND playlist_id = ?"
		setLastSyncQuery = "UPDATE account_synced_playlist_settings SET last_sync_description = ? WHERE account_id = ? AND playlist_id = ?"
	)
	return newBoolSyncSetting(e.HSyncDescription, func(val bool) error {
		e.HSyncDescription = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	}, e.HLastSyncDescription, func(val int64) error {
		e.HLastSyncDescription = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	})
}

func (e *AccountSyncedPlaylistSettings) Visibility() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_synced_playlist_settings SET sync_visibility = ? WHERE account_id = ? AND playlist_id = ?"
		setLastSyncQuery = "UPDATE account_synced_playlist_settings SET last_sync_visibility = ? WHERE account_id = ? AND playlist_id = ?"
	)
	return newBoolSyncSetting(e.HSyncVisibility, func(val bool) error {
		e.HSyncVisibility = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	}, e.HLastSyncVisibility, func(val int64) error {
		e.HLastSyncVisibility = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	})
}

func (e *AccountSyncedPlaylistSettings) Tracks() shared.SynchronizationSettings {
	const (
		setSyncQuery     = "UPDATE account_synced_playlist_settings SET sync_tracks = ? WHERE account_id = ? AND playlist_id = ?"
		setLastSyncQuery = "UPDATE account_synced_playlist_settings SET last_sync_tracks = ? WHERE account_id = ? AND playlist_id = ?"
	)
	return newBoolSyncSetting(e.HSyncTracks, func(val bool) error {
		e.HSyncTracks = val
		_, err := dbExec(context.Background(), setSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	}, e.HLastSyncTracks, func(val int64) error {
		e.HLastSyncTracks = val
		_, err := dbExec(context.Background(), setLastSyncQuery, val, e.HAccountID, e.HPlaylistID)
		return err
	})
}

func newBoolSyncSetting(
	sync bool,
	onSetSync func(bool) error,
	lastSync int64,
	onSetLastSync func(int64) error) *boolSyncSetting {
	return &boolSyncSetting{
		sync:          sync,
		onSetSync:     onSetSync,
		lastSync:      lastSync,
		onSetLastSync: onSetLastSync,
	}
}

type boolSyncSetting struct {
	sync      bool
	onSetSync func(newVal bool) error

	lastSync      int64
	onSetLastSync func(newVal int64) error
}

func (e boolSyncSetting) Synchronize() bool {
	return e.sync
}

func (e *boolSyncSetting) SetSynchronize(val bool) error {
	e.sync = val
	return e.onSetSync(val)
}

func (e boolSyncSetting) LastSynchronization() time.Time {
	return shared.TimeNano(e.lastSync)
}

func (e *boolSyncSetting) SetLastSynchronization(val time.Time) error {
	timestamp := shared.TimestampNano(val)
	return e.onSetLastSync(timestamp)
}
