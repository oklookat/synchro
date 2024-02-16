package syncerimpl

import (
	"context"
	"time"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

// Example: playlist names syncer per account.
func NewMetaAccount[T comparable](
	repo syncer.Repository[T],
	entityID shared.EntityID,
	lastSync time.Time,
	setLastSync func(time.Time) error,
	data T,
	setData func(context.Context, T) error,
) *MetaAccount[T] {
	if repo == nil || setLastSync == nil ||
		setData == nil {
		return nil
	}
	return &MetaAccount[T]{
		repo:     repo,
		entityID: entityID,

		lastSync:    lastSync,
		setLastSync: setLastSync,

		data:    data,
		setData: setData,
	}
}

type MetaAccount[T comparable] struct {
	// Example: playlist names syncer repo.
	repo syncer.Repository[T]

	// Example: playlist entity ID.
	entityID shared.EntityID

	// Example: playlist name last sync.
	lastSync time.Time

	// Example: set playlist last sync.
	setLastSync func(time.Time) error

	// Example: playlist name.
	data T

	// Example: set playlist name.
	setData func(context.Context, T) error
}

// Start syncing.
func (e *MetaAccount[T]) Start(ctx context.Context) error {
	remoteData := map[shared.EntityID]syncer.SyncableEntity[T]{
		e.entityID: &metaSyncableEntity[T]{
			id:      e.entityID,
			account: e,
			data:    e.data,
		},
	}
	sync := syncer.New(e.repo, remoteData, e.lastSync)
	if err := sync.FromRemote(ctx); err != nil {
		return err
	}
	return e.setLastSync(time.Now())
}

type metaSyncableEntity[T comparable] struct {
	id shared.EntityID

	account *MetaAccount[T]
	data    T
}

func (e metaSyncableEntity[T]) Get() T {
	return e.data
}

func (e *metaSyncableEntity[T]) Set(ctx context.Context, val T) error {
	e.data = val
	return e.account.setData(ctx, val)
}
