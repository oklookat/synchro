package syncerimpl

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/shared"
)

type likeableSyncableParam struct {
	id  shared.EntityID
	get bool

	likes   map[shared.EntityID]shared.RemoteID
	unliked map[shared.EntityID]shared.RemoteID
}

func (e likeableSyncableParam) Get() bool {
	return e.get
}

func (e *likeableSyncableParam) Set(ctx context.Context, val bool) error {
	e.get = val
	if val {
		// Because we working with likes/unlikes (add/delete).
		return errors.New("likeableSyncableParam: Set() cannot be true")
	}
	_, ok := e.likes[e.id]
	if !ok {
		return nil
	}
	if e.unliked != nil {
		e.unliked[e.id] = e.likes[e.id]
	}
	return nil
}
