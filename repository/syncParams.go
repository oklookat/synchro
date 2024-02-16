package repository

import (
	"context"

	"github.com/oklookat/synchro/shared"
)

type albumSyncParamIsSynced struct {
	origin *SyncedAlbum
}

func (e albumSyncParamIsSynced) Get() bool {
	return e.origin.HIsSynced
}

func (e *albumSyncParamIsSynced) Set(ctx context.Context, val bool) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_album SET is_synced=?,is_synced_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HIsSynced = val
		e.origin.HIsSyncedModifiedAt = now
	}
	return err
}

type artistSyncParamIsSynced struct {
	origin *SyncedArtist
}

func (e artistSyncParamIsSynced) Get() bool {
	return e.origin.HIsSynced
}

func (e *artistSyncParamIsSynced) Set(ctx context.Context, val bool) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_artist SET is_synced=?,is_synced_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HIsSynced = val
		e.origin.HIsSyncedModifiedAt = now
	}
	return err
}

type trackSyncParamIsSynced struct {
	origin *SyncedTrack
}

func (e trackSyncParamIsSynced) Get() bool {
	return e.origin.HIsSynced
}

func (e *trackSyncParamIsSynced) Set(ctx context.Context, val bool) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_track SET is_synced=?,is_synced_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HIsSynced = val
		e.origin.HIsSyncedModifiedAt = now
	}
	return err
}

type playlistSyncParamIsSynced struct {
	origin *SyncedPlaylist
}

func (e playlistSyncParamIsSynced) Get() bool {
	return e.origin.HIsSynced
}

func (e *playlistSyncParamIsSynced) Set(ctx context.Context, val bool) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_playlist SET is_synced=?,is_synced_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HIsSynced = val
		e.origin.HIsSyncedModifiedAt = now
	}
	return err
}

type playlistSyncParamName struct {
	origin *SyncedPlaylist
}

func (e playlistSyncParamName) Get() string {
	return e.origin.HName
}

func (e *playlistSyncParamName) Set(ctx context.Context, val string) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_playlist SET name=?,name_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HName = val
		e.origin.HNameModifiedAt = now
	}
	return err
}

type playlistSyncParamIsVisible struct {
	origin *SyncedPlaylist
}

func (e playlistSyncParamIsVisible) Get() bool {
	return e.origin.HIsVisible
}

func (e *playlistSyncParamIsVisible) Set(ctx context.Context, val bool) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_playlist SET is_visible=?,is_visible_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HIsVisible = val
		e.origin.HIsVisibleModifiedAt = now
	}
	return err
}

type playlistSyncParamDescription struct {
	origin *SyncedPlaylist
}

func (e playlistSyncParamDescription) Get() string {
	return e.origin.HDescription
}

func (e *playlistSyncParamDescription) Set(ctx context.Context, val string) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_playlist SET description=?,description_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HDescription = val
		e.origin.HDescriptionModifiedAt = now
	}
	return err
}

type playlistTrackSyncParamIsSynced struct {
	origin *SyncedPlaylistTrack
}

func (e playlistTrackSyncParamIsSynced) Get() bool {
	return e.origin.HIsSynced
}

func (e *playlistTrackSyncParamIsSynced) Set(ctx context.Context, val bool) error {
	now := shared.TimestampNanoNow()
	const query = "UPDATE synced_playlist_track SET is_synced=?,is_synced_modified_at=? WHERE id=?"
	_, err := dbExec(ctx, query, val, now, e.origin.HID)
	if err == nil {
		e.origin.HIsSynced = val
		e.origin.HIsSyncedModifiedAt = now
	}
	return err
}
