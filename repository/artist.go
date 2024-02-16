package repository

import (
	"context"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

// EntityRepository.
type EntityArtist struct {
}

func (e EntityArtist) CreateEntity() (shared.EntityID, error) {
	const query = `INSERT INTO artist DEFAULT VALUES RETURNING *`
	ent, err := dbGetOne[ArtistEntity](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return shared.EntityID(ent.ID()), err
}

func (e EntityArtist) DeleteNotLinked() error {
	const query = `DELETE FROM artist
WHERE NOT EXISTS (
	SELECT 1 FROM linked_artist
	WHERE linked_artist.artist_id = artist.id AND linked_artist.id_on_remote IS NOT NULL
) AND 
EXISTS (
	SELECT 1 FROM linked_artist
	WHERE linked_artist.artist_id = artist.id AND linked_artist.id_on_remote IS NULL
);`
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityArtist) DeleteAll() error {
	const query = "DELETE FROM artist"
	_, err := dbExec(context.Background(), query)
	return err
}

type ArtistEntity struct {
	HID uint64 `db:"id"`
}

func (e ArtistEntity) ID() uint64 {
	return e.HID
}

// Linkable.
func NewLinkableArtist(rem *Remote) *LinkableArtist {
	return &LinkableArtist{
		rem: rem,
	}
}

type LinkableArtist struct {
	rem *Remote
}

func (e LinkableArtist) CreateLink(ctx context.Context, eId shared.EntityID, id *shared.RemoteID) (linker.Linked, error) {
	const query = `INSERT INTO linked_artist (artist_id, remote_name, id_on_remote, modified_at)
	VALUES (?, ?, ?, ?) RETURNING *;`
	return dbGetOne[LinkedArtist](ctx, query, eId, e.rem.Name(), id, shared.TimestampNow())
}

func (e LinkableArtist) LinkedEntity(eId shared.EntityID) (linker.Linked, error) {
	const query = "SELECT * FROM linked_artist WHERE artist_id=? AND remote_name=? LIMIT 1"
	return dbGetOne[LinkedArtist](context.Background(), query, eId, e.rem.Name())
}

func (e LinkableArtist) LinkedRemoteID(id shared.RemoteID) (linker.Linked, error) {
	const query = "SELECT * FROM linked_artist WHERE id_on_remote=? AND remote_name=? LIMIT 1"
	return dbGetOne[LinkedArtist](context.Background(), query, id, e.rem.Name())
}

func (e LinkableArtist) Links() ([]linker.Linked, error) {
	const query = "SELECT * FROM linked_artist WHERE remote_name=?"
	return dbGetManyConvert[LinkedArtist, linker.Linked](context.Background(), nil, query, e.rem.Name())
}

func (e LinkableArtist) NotMatchedCount() (int, error) {
	query := getNotMatchedCountQuery("linked_artist", e.rem.Name())
	result, err := dbGetOne[int](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return *result, err
}

func DebugSetArtistMissing(artistID uint64, remoteName shared.RemoteName) error {
	const query = `UPDATE linked_artist SET id_on_remote=? WHERE artist_id=? AND remote_name=?`
	_, err := dbExec(context.Background(), query, artistID, remoteName)
	return err
}

// Linked.
type LinkedArtist struct {
	HID         uint64            `db:"id"`
	ArtistID    uint64            `db:"artist_id"`
	HRemoteName shared.RemoteName `db:"remote_name"`
	IdOnRemote  *shared.RemoteID  `db:"id_on_remote"`
	HModifiedAt int64             `db:"modified_at"`
}

func (e LinkedArtist) EntityID() shared.EntityID {
	return shared.EntityID(e.ArtistID)
}

func (e LinkedArtist) RemoteID() *shared.RemoteID {
	return e.IdOnRemote
}

func (e *LinkedArtist) SetRemoteID(id *shared.RemoteID) error {
	var copied *shared.RemoteID
	if id != nil {
		cp := *id
		copied = &cp
	}
	now := shared.TimestampNow()
	const query = "UPDATE linked_artist SET id_on_remote=?,modified_at=? WHERE id=?"
	_, err := dbExec(context.Background(), query, copied, now, e.HID)
	if err == nil {
		e.IdOnRemote = id
		e.HModifiedAt = now
	}
	return err
}

func (e LinkedArtist) ModifiedAt() time.Time {
	return shared.Time(e.HModifiedAt)
}

// Likeable.
type SyncablesArtist struct {
}

func (e SyncablesArtist) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	const query = `INSERT INTO synced_artist (artist_id) VALUES (?) RETURNING *`
	return dbGetOne[SyncedArtist](ctx, query, id)
}

func (e SyncablesArtist) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	const query = "SELECT * FROM synced_artist WHERE artist_id=? LIMIT 1"
	return dbGetOne[SyncedArtist](context.Background(), query, id)
}

func (e SyncablesArtist) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncablesArtist) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncablesArtist) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	const tableName = "synced_artist"
	const columnName = "is_synced"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}
	_, err := dbGetManyConvert[SyncedArtist, syncer.Synced[bool]](context.Background(), func(ar *SyncedArtist) error {
		result[shared.EntityID(ar.ArtistID)] = ar
		return nil
	}, query)
	return result, err
}

func (e SyncablesArtist) DeleteUnsynced() error {
	const query = "DELETE FROM synced_artist WHERE is_synced = 0"
	_, err := dbExec(context.Background(), query)
	return err
}

// Liked.
type SyncedArtist struct {
	HID                 uint64 `db:"id"`
	ArtistID            uint64 `db:"artist_id"`
	HIsSynced           bool   `db:"is_synced"`
	HIsSyncedModifiedAt int64  `db:"is_synced_modified_at"`
}

func (e *SyncedArtist) SyncableParam() syncer.SyncableParam[bool] {
	return &artistSyncParamIsSynced{
		origin: e,
	}
}

func (e SyncedArtist) ModifiedAt() time.Time {
	return shared.TimeNano(e.HIsSyncedModifiedAt)
}
