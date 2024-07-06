package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
)

// Linkable.
type LinkableEntity struct {
	entityName EntityName
	remoteName shared.RemoteName
}

func NewLinkableEntity(entityName EntityName, remoteName shared.RemoteName) *LinkableEntity {
	return &LinkableEntity{
		entityName: entityName,
		remoteName: remoteName,
	}
}

func (e LinkableEntity) CreateLink(ctx context.Context, eId shared.EntityID, id *shared.RemoteID) (linker.Linked, error) {
	query := fmt.Sprintf(`INSERT INTO linked_%s (id, entity_id, remote_name, id_on_remote, modified_at)
	VALUES (?, ?, ?, ?, ?) RETURNING *;`, e.entityName)
	return e.getOne(ctx, query, genRepositoryID(), eId, e.remoteName, id, shared.TimestampNow())
}

func (e LinkableEntity) LinkedEntity(eId shared.EntityID) (linker.Linked, error) {
	query := fmt.Sprintf("SELECT * FROM linked_%s WHERE entity_id=? AND remote_name=? LIMIT 1", e.entityName)
	return e.getOne(context.Background(), query, eId, e.remoteName)
}

func (e LinkableEntity) LinkedRemoteID(id shared.RemoteID) (linker.Linked, error) {
	query := fmt.Sprintf("SELECT * FROM linked_%s WHERE id_on_remote=? AND remote_name=? LIMIT 1", e.entityName)
	return e.getOne(context.Background(), query, id, e.remoteName)
}

func (e LinkableEntity) getOne(ctx context.Context, query string, args ...interface{}) (*LinkedEntity, error) {
	res, err := dbGetOne[LinkedEntity](ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if shared.IsNil(res) {
		return nil, err
	}
	res.entityName = e.entityName
	return res, err
}

func DebugSetEntityMissing(entityID uint64, entityName string, remoteName shared.RemoteName) error {
	query := fmt.Sprintf(`UPDATE linked_%s SET id_on_remote=? WHERE %s_id=? AND remote_name=?`, entityName, entityName)
	_, err := dbExec(context.Background(), query, entityID, remoteName)
	return err
}

// Linked.
type LinkedEntity struct {
	HID         shared.RepositoryID `db:"id"`
	HEntityID   shared.EntityID     `db:"entity_id"`
	HRemoteName shared.RemoteName   `db:"remote_name"`
	IdOnRemote  *shared.RemoteID    `db:"id_on_remote"`
	HModifiedAt int64               `db:"modified_at"`

	entityName EntityName `json:"-" db:"-"`
}

func (e LinkedEntity) EntityID() shared.EntityID {
	return shared.EntityID(e.HEntityID)
}

func (e LinkedEntity) RemoteID() *shared.RemoteID {
	return e.IdOnRemote
}

func (e *LinkedEntity) SetRemoteID(id *shared.RemoteID) error {
	var copied *shared.RemoteID
	if id != nil {
		cp := *id
		copied = &cp
	}
	now := shared.TimestampNow()
	query := fmt.Sprintf("UPDATE linked_%s SET id_on_remote=?,modified_at=? WHERE id=?", e.entityName)
	_, err := dbExec(context.Background(), query, copied, now, e.HID)
	if err == nil {
		e.IdOnRemote = id
		e.HModifiedAt = now
	}
	return err
}

func (e LinkedEntity) ModifiedAt() time.Time {
	return shared.Time(e.HModifiedAt)
}
