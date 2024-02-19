package repository

import (
	"context"
	"fmt"

	"github.com/oklookat/synchro/shared"
)

type EntityName string

func (e EntityName) String() string {
	return string(e)
}

type EntityRepository struct {
	name EntityName
}

// Example: artist.
func NewEntityRepository(name EntityName) EntityRepository {
	return EntityRepository{
		name: name,
	}
}

func (e EntityRepository) CreateEntity() (shared.EntityID, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id) RETURNING *`, e.name)
	ent, err := dbGetOne[Entity](context.Background(), query, genEntityID())
	if err != nil {
		return "", err
	}
	return shared.EntityID(ent.ID()), err
}

func (e EntityRepository) DeleteNotLinked() error {
	query := fmt.Sprintf(`DELETE FROM %s
	WHERE NOT EXISTS (
		SELECT 1 FROM linked_%s
		WHERE linked_%s.%s_id = %s.id AND linked_%s.id_on_remote IS NOT NULL
	) AND 
	EXISTS (
		SELECT 1 FROM linked_%s
		WHERE linked_%s.%s_id = %s.id AND linked_%s.id_on_remote IS NULL
	);`, e.name, e.name, e.name, e.name, e.name, e.name, e.name, e.name, e.name, e.name, e.name)
	_, err := dbExec(context.Background(), query)
	return err
}

func (e EntityRepository) DeleteAll() error {
	query := "DELETE FROM " + e.name
	_, err := dbExec(context.Background(), query.String())
	return err
}

type Entity struct {
	HID shared.EntityID `db:"id"`
}

func (e Entity) ID() shared.EntityID {
	return e.HID
}
