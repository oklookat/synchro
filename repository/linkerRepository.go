package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

func newEntityRepository(entityTableName, linkedTableName string) EntityRepository {
	return EntityRepository{
		entityTableName: entityTableName,
		linkedTableName: linkedTableName,
	}
}

type EntityRepository struct {
	entityTableName, linkedTableName string
}

func (e EntityRepository) CreateEntity() (streaming.DatabaseEntityID, error) {
	query := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES RETURNING *", e.entityTableName)
	ent, err := dbGetOne[Entity](context.Background(), query)
	if err != nil {
		return 0, err
	}
	return streaming.DatabaseEntityID(ent.ID()), err
}

func (e EntityRepository) DeleteNotLinked() error {
	query := fmt.Sprintf(`DELETE FROM %s
	WHERE NOT EXISTS (
		SELECT 1 FROM %s
		WHERE %s.entity_id = %s.id AND %s.id_on_service IS NOT NULL
	) AND 
	EXISTS (
		SELECT 1 FROM %s
		WHERE %s.entity_id = %s.id AND %s.id_on_service IS NULL
	);`, e.entityTableName, e.linkedTableName, e.linkedTableName, e.entityTableName,
		e.linkedTableName, e.linkedTableName, e.linkedTableName, e.entityTableName, e.linkedTableName)
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityRepository) DeleteAll() error {
	query := "DELETE FROM " + e.entityTableName
	_, err := dbExec(context.Background(), query)
	return err
}

type Entity struct {
	HID uint64 `db:"id"`
}

func (e Entity) ID() uint64 {
	return e.HID
}

func newLinkableEntity(linkedTableName string, rem *Service) *LinkableEntity {
	return &LinkableEntity{
		rem:             rem,
		linkedTableName: linkedTableName,
	}
}

type LinkableEntity struct {
	rem             *Service
	linkedTableName string
}

func (e LinkableEntity) CreateLink(ctx context.Context, eId streaming.DatabaseEntityID, id *streaming.ServiceEntityID) (linker.Linked, error) {
	query := fmt.Sprintf(`INSERT INTO %s (entity_id, service_name, id_on_service, modified_at)
	VALUES (?, ?, ?, ?) RETURNING *`, e.linkedTableName)
	res, err := dbGetOne[LinkedEntity](ctx, query, eId, e.rem.Name(), id, shared.TimestampNow())
	if err != nil {
		return nil, err
	}
	res.linkedTableName = e.linkedTableName
	return res, err
}

func (e LinkableEntity) LinkedEntity(eId streaming.DatabaseEntityID) (linker.Linked, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE entity_id=? AND service_name=? LIMIT 1", e.linkedTableName)
	res, err := dbGetOne[LinkedEntity](context.Background(), query, eId, e.rem.Name())
	if err != nil {
		return nil, err
	}
	res.linkedTableName = e.linkedTableName
	return res, err
}

func (e LinkableEntity) LinkedRemoteID(id streaming.ServiceEntityID) (linker.Linked, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id_on_service=? AND service_name=? LIMIT 1", e.linkedTableName)
	res, err := dbGetOne[LinkedEntity](context.Background(), query, id, e.rem.Name())
	if err != nil {
		return nil, err
	}
	res.linkedTableName = e.linkedTableName
	return res, err
}

// Linked.
type LinkedEntity struct {
	HID         uint64                     `db:"id"`
	HEntityID   uint64                     `db:"entity_id"`
	HRemoteName streaming.ServiceName      `db:"service_name"`
	IdOnRemote  *streaming.ServiceEntityID `db:"id_on_service"`
	HModifiedAt int64                      `db:"modified_at"`

	linkedTableName string `db:"-" json:"-"`
}

func (e LinkedEntity) EntityID() streaming.DatabaseEntityID {
	return streaming.DatabaseEntityID(e.HEntityID)
}

func (e LinkedEntity) ServiceEntityID() *streaming.ServiceEntityID {
	return e.IdOnRemote
}

func (e *LinkedEntity) SetServiceEntityID(id *streaming.ServiceEntityID) error {
	var copied *streaming.ServiceEntityID
	if id != nil {
		cp := *id
		copied = &cp
	}
	now := shared.TimestampNow()
	query := fmt.Sprintf("UPDATE %s SET id_on_service=?,modified_at=? WHERE id=?", e.linkedTableName)
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
