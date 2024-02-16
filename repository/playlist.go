package repository

import (
	"context"
	"errors"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

// EntityRepository.
type EntityPlaylist struct {
}

func (e EntityPlaylist) CreateEntity() (shared.EntityID, error) {
	const query = `INSERT INTO playlist DEFAULT VALUES RETURNING *`
	ent, err := dbGetOne[PlaylistEntity](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return shared.EntityID(ent.ID()), err
}

func (e EntityPlaylist) DeleteNotLinked() error {
	const query = `DELETE FROM playlist
WHERE NOT EXISTS (
	SELECT 1 FROM linked_playlist
	WHERE linked_playlist.playlist_id = playlist.id AND linked_playlist.id_on_remote IS NOT NULL
) AND 
EXISTS (
	SELECT 1 FROM linked_playlist
	WHERE linked_playlist.playlist_id = playlist.id AND linked_playlist.id_on_remote IS NULL
);`
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityPlaylist) DeleteAll() error {
	const query = "DELETE FROM playlist"
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityPlaylist) DeleteEntity(id shared.EntityID) error {
	const query = "DELETE FROM playlist WHERE id = ?"
	_, err := dbExec(context.Background(), query, id)
	return err
}

func playlistEntityByID(ctx context.Context, id shared.EntityID) (*PlaylistEntity, error) {
	const query = "SELECT * FROM playlist WHERE id = ? LIMIT 1"
	return dbGetOne[PlaylistEntity](ctx, query, id)
}

type PlaylistEntity struct {
	HID uint64 `db:"id" json:"id"`
}

func (e PlaylistEntity) ID() uint64 {
	return e.HID
}

// Linkable.
func NewLinkablePlaylist(account *Account) *LinkablePlaylist {
	return &LinkablePlaylist{
		account: account,
	}
}

type LinkablePlaylist struct {
	account *Account
}

func (e LinkablePlaylist) CreateLink(ctx context.Context, eId shared.EntityID, id shared.RemoteID) (linker.LinkedDynamic, error) {
	const query = `INSERT INTO linked_playlist (playlist_id, account_id, id_on_remote, modified_at)
	VALUES (?, ?, ?, ?) RETURNING *;`
	return dbGetOne[LinkedPlaylist](ctx, query, eId, e.account.ID(), id, shared.TimestampNow())
}

func (e LinkablePlaylist) LinkedEntity(eId shared.EntityID) (linker.LinkedDynamic, error) {
	const query = "SELECT * FROM linked_playlist WHERE playlist_id=? AND account_id=? LIMIT 1"
	return dbGetOne[LinkedPlaylist](context.Background(), query, eId, e.account.ID())
}

func (e LinkablePlaylist) LinkedRemoteID(id shared.RemoteID) (linker.LinkedDynamic, error) {
	const query = "SELECT * FROM linked_playlist WHERE id_on_remote=? AND account_id=?"
	return dbGetOne[LinkedPlaylist](context.Background(), query, id, e.account.ID())
}

// Linked.
type LinkedPlaylist struct {
	HID         uint64          `db:"id"`
	HPlaylistID uint64          `db:"playlist_id"`
	HAccountID  uint64          `db:"account_id"`
	IdOnRemote  shared.RemoteID `db:"id_on_remote"`
	HModifiedAt int64           `db:"modified_at"`
}

func (e LinkedPlaylist) EntityID() shared.EntityID {
	return shared.EntityID(e.HPlaylistID)
}

func (e LinkedPlaylist) RemoteID() shared.RemoteID {
	return e.IdOnRemote
}

func (e LinkedPlaylist) ModifiedAt() time.Time {
	return shared.Time(e.HModifiedAt)
}

// Playlist.
func NewSyncablesPlaylistIsSynced() syncer.Repository[bool] {
	return &SyncablesPlaylist{}
}

// Delete / add.
type SyncablesPlaylist struct {
}

func (e SyncablesPlaylist) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	const query = `INSERT INTO synced_playlist (playlist_id) VALUES (?) RETURNING *`
	return dbGetOne[SyncedPlaylist](ctx, query, id)
}

func (e SyncablesPlaylist) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	const query = "SELECT * FROM synced_playlist WHERE playlist_id=? LIMIT 1"
	return dbGetOne[SyncedPlaylist](context.Background(), query, id)
}

func (e SyncablesPlaylist) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesPlaylist) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesPlaylist) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	const tableName = "synced_playlist"
	const columnName = "is_synced"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}
	_, err := dbGetManyConvert[SyncedPlaylist, syncer.Synced[bool]](context.Background(), func(pl *SyncedPlaylist) error {
		result[shared.EntityID(pl.HPlaylistID)] = pl
		return nil
	}, query)
	return result, err
}

func (e SyncablesPlaylist) DeleteUnsynced() error {
	const query = `DELETE FROM playlist WHERE id IN (
SELECT playlist_id FROM synced_playlist WHERE is_synced = 0);`
	_, err := dbExec(context.Background(), query)
	return err
}

func syncedPlaylistByPlaylistID(ctx context.Context, playlistID shared.EntityID) (*SyncedPlaylist, error) {
	const query = "SELECT * FROM synced_playlist WHERE playlist_id = ? LIMIT 1"
	return dbGetOne[SyncedPlaylist](ctx, query, playlistID)
}

type SyncedPlaylist struct {
	HID         uint64 `db:"id"`
	HPlaylistID uint64 `db:"playlist_id"`

	HIsSynced           bool  `db:"is_synced"`
	HIsSyncedModifiedAt int64 `db:"is_synced_modified_at"`

	HIsVisible           bool  `db:"is_visible"`
	HIsVisibleModifiedAt int64 `db:"is_visible_modified_at"`

	HName           string `db:"name"`
	HNameModifiedAt int64  `db:"name_modified_at"`

	HDescription           string `db:"description"`
	HDescriptionModifiedAt int64  `db:"description_modified_at"`
}

func (e *SyncedPlaylist) SyncableParam() syncer.SyncableParam[bool] {
	return &playlistSyncParamIsSynced{
		origin: e,
	}
}

func (e SyncedPlaylist) ModifiedAt() time.Time {
	return shared.TimeNano(e.HIsSyncedModifiedAt)
}

func NewSyncablesPlaylistTrack(ctx context.Context, playlistEntity shared.EntityID) (*SyncablesPlaylistTrack, error) {
	ent, err := playlistEntityByID(ctx, playlistEntity)
	if err != nil {
		return nil, err
	}
	syncd, err := syncedPlaylistByPlaylistID(ctx, shared.EntityID(ent.ID()))
	if err != nil {
		return nil, err
	}
	return &SyncablesPlaylistTrack{
		SyncedPlaylistID: syncd.HID,
	}, err
}

// Playlist tracks.
type SyncablesPlaylistTrack struct {
	SyncedPlaylistID uint64
}

func (e SyncablesPlaylistTrack) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	const query = `INSERT INTO synced_playlist_track (synced_playlist_id, track_id, is_synced_modified_at) VALUES (?, ?, ?) RETURNING *;`
	return dbGetOne[SyncedPlaylistTrack](ctx, query, e.SyncedPlaylistID, id, shared.TimestampNanoNow())
}

func (e SyncablesPlaylistTrack) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	const query = "SELECT * FROM synced_playlist_track WHERE synced_playlist_id=? AND track_id=? LIMIT 1"
	return dbGetOne[SyncedPlaylistTrack](context.Background(), query, e.SyncedPlaylistID, id)
}

func (e SyncablesPlaylistTrack) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesPlaylistTrack) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesPlaylistTrack) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	const tableName = "synced_playlist_track"
	const columnName = "is_synced"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName) + " AND synced_playlist_id = ?"
	} else {
		query = genOlderQuery(tableName, date, columnName) + " AND synced_playlist_id = ?"
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}
	_, err := dbGetManyConvert[SyncedPlaylistTrack, syncer.Synced[bool]](context.Background(), func(tr *SyncedPlaylistTrack) error {
		result[shared.EntityID(tr.HTrackID)] = tr
		return nil
	}, query, e.SyncedPlaylistID)
	return result, err
}

func (e SyncablesPlaylistTrack) DeleteUnsynced() error {
	const query = "DELETE FROM synced_playlist_track WHERE is_synced = 0 AND synced_playlist_id=?"
	_, err := dbExec(context.Background(), query, e.SyncedPlaylistID)
	return err
}

// Liked.
type SyncedPlaylistTrack struct {
	HID               uint64 `db:"id" json:"id"`
	HSyncedPlaylistID uint64 `db:"synced_playlist_id" json:"syncedPlaylistID"`
	HTrackID          uint64 `db:"track_id" json:"trackID"`

	HIsSynced           bool  `db:"is_synced" json:"isSynced"`
	HIsSyncedModifiedAt int64 `db:"is_synced_modified_at" json:"isSyncedModifiedAt"`
}

func (e *SyncedPlaylistTrack) SyncableParam() syncer.SyncableParam[bool] {
	return &playlistTrackSyncParamIsSynced{
		origin: e,
	}
}

func (e SyncedPlaylistTrack) ModifiedAt() time.Time {
	return shared.TimeNano(e.HIsSyncedModifiedAt)
}

// Sync playlist vis.
func NewSyncablesPlaylistIsVisible() syncer.Repository[bool] {
	return &SyncablesPlaylistIsVisible{}
}

type SyncablesPlaylistIsVisible struct {
	repo SyncablesPlaylist
}

func (e SyncablesPlaylistIsVisible) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	// Not used.
	return nil, nil
}

func (e SyncablesPlaylistIsVisible) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	syncedPlaylist, err := e.repo.Synced(id)
	if err != nil {
		return nil, err
	}
	real, ok := syncedPlaylist.(*SyncedPlaylist)
	if !ok {
		return nil, errors.New("synced real, ok := syncedPlaylist.(*SyncedPlaylist)")
	}
	return &SyncedPlaylistIsVisible{
		playlist: real,
	}, nil
}

func (e SyncablesPlaylistIsVisible) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesPlaylistIsVisible) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesPlaylistIsVisible) DeleteUnsynced() error {
	return nil
}

func (e SyncablesPlaylistIsVisible) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	const tableName = "synced_playlist"
	const columnName = "is_visible"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}
	_, err := dbGetMany(context.Background(), query, func(pl *SyncedPlaylist) error {
		result[shared.EntityID(pl.HPlaylistID)] = SyncedPlaylistIsVisible{playlist: pl}
		return nil
	})
	return result, err
}

type SyncedPlaylistIsVisible struct {
	playlist *SyncedPlaylist
}

func (e SyncedPlaylistIsVisible) SyncableParam() syncer.SyncableParam[bool] {
	return &playlistSyncParamIsVisible{
		origin: e.playlist,
	}
}

func (e SyncedPlaylistIsVisible) ModifiedAt() time.Time {
	return shared.TimeNano(e.playlist.HIsVisibleModifiedAt)
}

// Sync playlist name.
func NewSyncablesPlaylistName() syncer.Repository[string] {
	return SyncablesPlaylistName{}
}

type SyncablesPlaylistName struct {
	repo SyncablesPlaylist
}

func (e SyncablesPlaylistName) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[string], error) {
	// Not used.
	return nil, nil
}

func (e SyncablesPlaylistName) Synced(id shared.EntityID) (syncer.Synced[string], error) {
	syncedPlaylist, err := e.repo.Synced(id)
	if err != nil {
		return nil, err
	}
	real, ok := syncedPlaylist.(*SyncedPlaylist)
	if !ok {
		return nil, errors.New("synced real, ok := syncedPlaylist.(*SyncedPlaylist)")
	}
	return &SyncedPlaylistName{
		playlist: real,
	}, nil
}

func (e SyncablesPlaylistName) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[string], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesPlaylistName) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[string], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesPlaylistName) DeleteUnsynced() error {
	return nil
}

func (e SyncablesPlaylistName) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[string], error) {
	const tableName = "synced_playlist"
	const columnName = "name"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[string]{}
	_, err := dbGetMany(context.Background(), query, func(pl *SyncedPlaylist) error {
		result[shared.EntityID(pl.HPlaylistID)] = SyncedPlaylistName{playlist: pl}
		return nil
	})
	return result, err
}

type SyncedPlaylistName struct {
	playlist *SyncedPlaylist
}

func (e SyncedPlaylistName) SyncableParam() syncer.SyncableParam[string] {
	return &playlistSyncParamName{
		origin: e.playlist,
	}
}

func (e SyncedPlaylistName) ModifiedAt() time.Time {
	return shared.TimeNano(e.playlist.HNameModifiedAt)
}

// Sync playlist desc.
func NewSyncablesPlaylistDescription() syncer.Repository[string] {
	return SyncablesPlaylistDescription{}
}

type SyncablesPlaylistDescription struct {
	repo SyncablesPlaylist
}

func (e SyncablesPlaylistDescription) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[string], error) {
	// Not used.
	return nil, nil
}

func (e SyncablesPlaylistDescription) Synced(id shared.EntityID) (syncer.Synced[string], error) {
	syncedPlaylist, err := e.repo.Synced(id)
	if err != nil {
		return nil, err
	}
	real, ok := syncedPlaylist.(*SyncedPlaylist)
	if !ok {
		return nil, errors.New("synced real, ok := syncedPlaylist.(*SyncedPlaylist)")
	}
	return &SyncedPlaylistDescription{
		playlist: real,
	}, nil
}

func (e SyncablesPlaylistDescription) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[string], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesPlaylistDescription) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[string], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesPlaylistDescription) DeleteUnsynced() error {
	return nil
}

func (e SyncablesPlaylistDescription) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[string], error) {
	const tableName = "synced_playlist"
	const columnName = "description"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[string]{}
	_, err := dbGetMany(context.Background(), query, func(pl *SyncedPlaylist) error {
		result[shared.EntityID(pl.HPlaylistID)] = SyncedPlaylistName{playlist: pl}
		return nil
	})
	return result, err
}

type SyncedPlaylistDescription struct {
	playlist *SyncedPlaylist
}

func (e SyncedPlaylistDescription) SyncableParam() syncer.SyncableParam[string] {
	return &playlistSyncParamDescription{
		origin: e.playlist,
	}
}

func (e SyncedPlaylistDescription) ModifiedAt() time.Time {
	return shared.TimeNano(e.playlist.HDescriptionModifiedAt)
}
