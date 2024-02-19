package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

// Syncable.
type SyncableEntity struct {
	name EntityName
}

// Example: artist.
func NewSyncableEntity(name EntityName) SyncableEntity {
	return SyncableEntity{
		name: name,
	}
}

func (e SyncableEntity) CreateSynced(ctx context.Context, id shared.EntityID) (syncer.Synced[bool], error) {
	query := fmt.Sprintf(`INSERT INTO synced_%s (id, %s_id) VALUES (?, ?) RETURNING *`, e.name, e.name)
	return e.getOne(context.Background(), query, genRepositoryID(), id)
}

func (e SyncableEntity) Synced(id shared.EntityID) (syncer.Synced[bool], error) {
	query := fmt.Sprintf("SELECT * FROM synced_%s WHERE %s_id=? LIMIT 1", e.name, e.name)
	return e.getOne(context.Background(), query, id)
}

func (e SyncableEntity) getOne(ctx context.Context, query string, args ...interface{}) (syncer.Synced[bool], error) {
	res, err := dbGetOne[SyncedEntity](context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	res.entityName = e.name
	return res, err
}

func (e SyncableEntity) SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, newerThan, true)
}

func (e SyncableEntity) SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]syncer.Synced[bool], error) {
	return e.syncesOlderNewer(ctx, olderThan, false)
}

func (e SyncableEntity) syncesOlderNewer(ctx context.Context, date time.Time, newerThan bool) (map[shared.EntityID]syncer.Synced[bool], error) {
	tableName := "synced_" + e.name.String()
	const columnName = "is_synced"

	var query string
	if newerThan {
		query = genNewerQuery(tableName, date, columnName)
	} else {
		query = genOlderQuery(tableName, date, columnName)
	}

	result := map[shared.EntityID]syncer.Synced[bool]{}

	_, err := dbGetManyConvert[SyncedEntity, syncer.Synced[bool]](context.Background(), func(ent *SyncedEntity) error {
		ent.entityName = e.name
		result[shared.EntityID(ent.HEntityID)] = ent
		return nil
	}, query)

	return result, err
}

func (e SyncableEntity) DeleteUnsynced() error {
	query := fmt.Sprintf("DELETE FROM synced_%s WHERE is_synced = 0", e.name)
	_, err := dbExec(context.Background(), query)
	return err
}

// Liked.
type SyncedEntity struct {
	HID                 shared.RepositoryID `db:"id"`
	HEntityID           shared.EntityID     `db:"entity_id"`
	HIsSynced           bool                `db:"is_synced"`
	HIsSyncedModifiedAt int64               `db:"is_synced_modified_at"`

	entityName EntityName `json:"-" db:"-"`
}

func (e *SyncedEntity) SyncableParam() syncer.SyncableParam[bool] {
	return newIsSyncedParam(e)
}

func (e SyncedEntity) ModifiedAt() time.Time {
	return shared.TimeNano(e.HIsSyncedModifiedAt)
}

type isSyncedParam struct {
	syncedEntity *SyncedEntity
}

func newIsSyncedParam(syncedEntity *SyncedEntity) *isSyncedParam {
	return &isSyncedParam{
		syncedEntity: syncedEntity,
	}
}

func (e isSyncedParam) Get() bool {
	return e.syncedEntity.HIsSynced
}

func (e *isSyncedParam) Set(ctx context.Context, val bool) error {
	nowNano := shared.TimestampNanoNow()
	query := fmt.Sprintf("UPDATE synced_%s SET is_synced=?,is_synced_modified_at=? WHERE id=?", e.syncedEntity.entityName)
	_, err := dbExec(ctx, query, val, nowNano, e.syncedEntity.HID)
	if err == nil {
		e.syncedEntity.HIsSynced = val
		e.syncedEntity.HIsSyncedModifiedAt = nowNano
	}
	return err
}
