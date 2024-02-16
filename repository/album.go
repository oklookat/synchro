package repository

import (
	"context"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

// EntityRepository.
type EntityAlbum struct {
}

func (e EntityAlbum) CreateEntity() (shared.EntityID, error) {
	const query = `INSERT INTO album DEFAULT VALUES RETURNING *`
	ent, err := dbGetOne[AlbumEntity](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return shared.EntityID(ent.ID()), err
}

func (e EntityAlbum) DeleteNotLinked() error {
	const query = `DELETE FROM album
WHERE NOT EXISTS (
	SELECT 1 FROM linked_album
	WHERE linked_album.album_id = album.id AND linked_album.id_on_remote IS NOT NULL
) AND 
EXISTS (
	SELECT 1 FROM linked_album
	WHERE linked_album.album_id = album.id AND linked_album.id_on_remote IS NULL
);`
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityAlbum) DeleteAll() error {
	const query = "DELETE FROM album"
	_, err := dbExec(context.Background(), query)
	return err
}

type AlbumEntity struct {
	HID uint64 `db:"id"`
}

func (e AlbumEntity) ID() uint64 {
	return e.HID
}

// Linkable.
func NewLinkableAlbum(rem *Remote) *LinkableAlbum {
	return &LinkableAlbum{
		rem: rem,
	}
}

type LinkableAlbum struct {
	rem *Remote
}

func (e LinkableAlbum) CreateLink(ctx context.Context, eId shared.EntityID, id *shared.RemoteID) (linker.Linked, error) {
	const query = `INSERT INTO linked_album (album_id, remote_name, id_on_remote, modified_at)
	VALUES (?, ?, ?, ?) RETURNING *;`
	return dbGetOne[LinkedAlbum](ctx, query, eId, e.rem.Name(), id, shared.TimestampNow())
}

func (e LinkableAlbum) LinkedEntity(eId shared.EntityID) (linker.Linked, error) {
	const query = "SELECT * FROM linked_album WHERE album_id=? AND remote_name=? LIMIT 1"
	return dbGetOne[LinkedAlbum](context.Background(), query, eId, e.rem.Name())
}

func (e LinkableAlbum) LinkedRemoteID(id shared.RemoteID) (linker.Linked, error) {
	const query = "SELECT * FROM linked_album WHERE id_on_remote=? AND remote_name=? LIMIT 1"
	return dbGetOne[LinkedAlbum](context.Background(), query, id, e.rem.Name())
}

func (e LinkableAlbum) Links() ([]linker.Linked, error) {
	const query = "SELECT * FROM linked_album WHERE remote_name=?"
	return dbGetManyConvert[LinkedAlbum, linker.Linked](context.Background(), nil, query, e.rem.Name())
}

func (e LinkableAlbum) NotMatchedCount() (int, error) {
	query := getNotMatchedCountQuery("linked_album", e.rem.Name())
	result, err := dbGetOne[int](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return *result, err
}

// Linked.
type LinkedAlbum struct {
	HID         uint64            `db:"id"`
	HAlbumID    uint64            `db:"album_id"`
	HRemoteName shared.RemoteName `db:"remote_name"`
	IdOnRemote  *shared.RemoteID  `db:"id_on_remote"`
	HModifiedAt int64             `db:"modified_at"`
}

func (e LinkedAlbum) EntityID() shared.EntityID {
	return shared.EntityID(e.HAlbumID)
}

func (e LinkedAlbum) RemoteID() *shared.RemoteID {
	return e.IdOnRemote
}

func (e *LinkedAlbum) SetRemoteID(id *shared.RemoteID) error {
	var copied *shared.RemoteID
	if id != nil {
		cp := *id
		copied = &cp
	}
	now := shared.TimestampNow()
	const query = "UPDATE linked_album SET id_on_remote=?,modified_at=? WHERE id=?"
	_, err := dbExec(context.Background(), query, copied, now, e.HID)
	if err == nil {
		e.IdOnRemote = id
		e.HModifiedAt = now
	}
	return err
}

func (e LinkedAlbum) ModifiedAt() time.Time {
	return shared.Time(e.HModifiedAt)
}

type SyncablesAlbum struct {
}

func (e SyncablesAlbum) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	const query = `INSERT INTO synced_album (album_id) VALUES (?) RETURNING *;`
	return dbGetOne[SyncedAlbum](ctx, query, id)
}

func (e SyncablesAlbum) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	const query = "SELECT * FROM synced_album WHERE album_id=? LIMIT 1"
	return dbGetOne[SyncedAlbum](context.Background(), query, id)
}

func (e SyncablesAlbum) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesAlbum) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesAlbum) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	const tableName = "synced_album"
	const columnName = "is_synced"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}
	_, err := dbGetManyConvert[SyncedAlbum, syncer.Synced[bool]](context.Background(), func(al *SyncedAlbum) error {
		result[shared.EntityID(al.HAlbumID)] = al
		return nil
	}, query)
	return result, err
}

func (e SyncablesAlbum) DeleteUnsynced() error {
	const query = "DELETE FROM synced_album WHERE is_synced = 0"
	_, err := dbExec(context.Background(), query)
	return err
}

// Liked.
type SyncedAlbum struct {
	HID                 uint64 `db:"id"`
	HAlbumID            uint64 `db:"album_id"`
	HIsSynced           bool   `db:"is_synced"`
	HIsSyncedModifiedAt int64  `db:"is_synced_modified_at"`
}

func (e *SyncedAlbum) SyncableParam() syncer.SyncableParam[bool] {
	return &albumSyncParamIsSynced{
		origin: e,
	}
}

func (e SyncedAlbum) ModifiedAt() time.Time {
	return shared.TimeNano(e.HIsSyncedModifiedAt)
}
