package deezer

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/oklookat/deezus/schema"
	"github.com/oklookat/synchro/shared"
)

func newEntity(id, name string) *Entity {
	return &Entity{
		id:   id,
		name: name,
	}
}

type Entity struct {
	id   string
	name string
}

func (e Entity) RemoteName() shared.RemoteName {
	return _repo.Name()
}

func (e Entity) ID() shared.RemoteID {
	return shared.RemoteID(e.id)
}

func (e Entity) Name() string {
	return e.name
}

func getCoverURL(urld string) *url.URL {
	return schema.GetPictureURL(urld, schema.PictureSizeMedium)
}

func isNotFound(err error) bool {
	deezErr := &schema.Error{}
	return errors.As(err, deezErr) && deezErr.Code == schema.ErrorCodeDataNotFound
}

func remoteToSchemaID(id shared.RemoteID) (schema.ID, error) {
	conv, err := strconv.ParseInt(id.String(), 10, 64)
	return schema.ID(conv), err
}
