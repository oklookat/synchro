package vkmusic

import (
	"errors"

	"github.com/oklookat/govkm/schema"
	"github.com/oklookat/synchro/streaming"
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

func (e Entity) ServiceName() streaming.ServiceName {
	return _repo.Name()
}

func (e Entity) ID() streaming.ServiceEntityID {
	return streaming.ServiceEntityID(e.id)
}

func (e Entity) Name() string {
	return e.name
}

func isNotFound(err error) bool {
	var respErr schema.ResponseError
	if errors.As(err, &respErr) {
		return respErr.IsNotFound()
	}
	return false
}

// Is track not official VKM track (UGC)?
func isUgcTrack(tr schema.Track) bool {
	return tr.IsUnofficial()
}
