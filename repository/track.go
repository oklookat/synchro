package repository

import (
	"context"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

// EntityRepository.
type EntityTrack struct {
}

func (e EntityTrack) CreateEntity() (shared.EntityID, error) {
	const query = `INSERT INTO track DEFAULT VALUES RETURNING *`
	ent, err := dbGetOne[TrackEntity](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return shared.EntityID(ent.ID()), err
}

func (e EntityTrack) DeleteNotLinked() error {
	const query = `DELETE FROM track
WHERE NOT EXISTS (
	SELECT 1 FROM linked_track
	WHERE linked_track.track_id = track.id AND linked_track.id_on_remote IS NOT NULL
) AND 
EXISTS (
	SELECT 1 FROM linked_track
	WHERE linked_track.track_id = track.id AND linked_track.id_on_remote IS NULL
);`
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityTrack) DeleteAll() error {
	const query = "DELETE FROM track"
	_, err := dbExec(context.Background(), query)
	return err
}

type TrackEntity struct {
	HID uint64 `db:"id"`
}

func (e TrackEntity) ID() uint64 {
	return e.HID
}

// Linkable.
func NewLinkableTrack(rem *Remote) *LinkableTrack {
	return &LinkableTrack{
		rem: rem,
	}
}

type LinkableTrack struct {
	rem *Remote
}

func (e LinkableTrack) CreateLink(ctx context.Context, eId shared.EntityID, id *shared.RemoteID) (linker.Linked, error) {
	const query = `INSERT INTO linked_track (track_id, remote_name, id_on_remote, modified_at)
	VALUES (?, ?, ?, ?) RETURNING *;`
	return dbGetOne[LinkedTrack](ctx, query, eId, e.rem.Name(), id, shared.TimestampNow())
}

func (e LinkableTrack) LinkedEntity(eId shared.EntityID) (linker.Linked, error) {
	const query = "SELECT * FROM linked_track WHERE track_id=? AND remote_name=? LIMIT 1"
	return dbGetOne[LinkedTrack](context.Background(), query, eId, e.rem.Name())
}

func (e LinkableTrack) LinkedRemoteID(id shared.RemoteID) (linker.Linked, error) {
	const query = "SELECT * FROM linked_track WHERE id_on_remote=? AND remote_name=? LIMIT 1"
	return dbGetOne[LinkedTrack](context.Background(), query, id, e.rem.Name())
}

func (e LinkableTrack) Links() ([]linker.Linked, error) {
	const query = "SELECT * FROM linked_track WHERE remote_name=?"
	linked, err := dbGetManyConvert[LinkedTrack, linker.Linked](context.Background(), nil, query, e.rem.Name())
	return linked, err
}
func (e LinkableTrack) NotMatchedCount() (int, error) {
	query := getNotMatchedCountQuery("linked_track", e.rem.Name())
	result, err := dbGetOne[int](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return *result, err
}

// Linked.
type LinkedTrack struct {
	HID         uint64            `db:"id"`
	HTrackID    uint64            `db:"track_id"`
	HRemoteName shared.RemoteName `db:"remote_name"`
	IdOnRemote  *shared.RemoteID  `db:"id_on_remote"`
	HModifiedAt int64             `db:"modified_at"`
}

func (e LinkedTrack) EntityID() shared.EntityID {
	return shared.EntityID(e.HTrackID)
}

func (e LinkedTrack) RemoteID() *shared.RemoteID {
	return e.IdOnRemote
}

func (e *LinkedTrack) SetRemoteID(id *shared.RemoteID) error {
	var copied *shared.RemoteID
	if id != nil {
		cp := *id
		copied = &cp
	}
	now := shared.TimestampNow()
	const query = "UPDATE linked_track SET id_on_remote=?,modified_at=? WHERE id=?"
	_, err := dbExec(context.Background(), query, copied, now, e.HID)
	if err == nil {
		e.IdOnRemote = id
		e.HModifiedAt = now
	}
	return err
}

func (e LinkedTrack) ModifiedAt() time.Time {
	return shared.Time(e.HModifiedAt)
}

// Likeable.
type SyncablesTrack struct {
}

func (e SyncablesTrack) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	const query = `INSERT INTO synced_track (track_id, is_synced_modified_at) VALUES (?, ?) RETURNING *`
	return dbGetOne[SyncedTrack](ctx, query, id, shared.TimestampNanoNow())
}

func (e SyncablesTrack) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	const query = "SELECT * FROM synced_track WHERE track_id=? LIMIT 1"
	return dbGetOne[SyncedTrack](context.Background(), query, id)
}

func (e SyncablesTrack) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesTrack) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesTrack) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	const tableName = "synced_track"
	const columnName = "is_synced"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}
	_, err := dbGetManyConvert[SyncedTrack, syncer.Synced[bool]](context.Background(), func(tr *SyncedTrack) error {
		result[shared.EntityID(tr.HTrackID)] = tr
		return nil
	}, query)
	return result, err
}

func (e SyncablesTrack) DeleteUnsynced() error {
	const query = "DELETE FROM synced_track WHERE is_synced = 0"
	_, err := dbExec(context.Background(), query)
	return err
}

// Liked.
type SyncedTrack struct {
	HID                 uint64 `db:"id"`
	HTrackID            uint64 `db:"track_id"`
	HIsSynced           bool   `db:"is_synced"`
	HIsSyncedModifiedAt int64  `db:"is_synced_modified_at"`
}

func (e *SyncedTrack) SyncableParam() syncer.SyncableParam[bool] {
	return &trackSyncParamIsSynced{
		origin: e,
	}
}

func (e SyncedTrack) ModifiedAt() time.Time {
	return shared.TimeNano(e.HIsSyncedModifiedAt)
}
