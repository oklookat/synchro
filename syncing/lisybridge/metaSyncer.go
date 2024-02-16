package lisybridge

import (
	"context"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
	"github.com/oklookat/synchro/syncing/syncerimpl"
)

type (
	// Syncable account.
	MetaAccount[T comparable] interface {
		// Get account playlists.
		MetaThings() []MetaThing[T]
	}

	// Example: playlist.
	MetaThing[T comparable] interface {
		// Example: get playlist name sync settings.
		SyncSetting() shared.SynchronizationSettings

		// Example: playlist entity ID.
		EntityID() shared.EntityID

		// Example: get playlist name.
		Param() T

		// Example: set playlist name.
		SetParam(context.Context, T) error
	}
)

func NewMetaSyncer[T comparable](repo syncer.Repository[T], accounts []MetaAccount[T]) *MetaSyncer[T] {
	return &MetaSyncer[T]{
		repo:     repo,
		accounts: accounts,
	}
}

// Example: playlists string syncer (name; description; etc).
type MetaSyncer[T comparable] struct {
	repo     syncer.Repository[T]
	accounts []MetaAccount[T]
}

func (e MetaSyncer[T]) Sync(ctx context.Context) error {
	if len(e.accounts) == 0 {
		return nil
	}
	return sync2stages(func() error {
		for _, account := range e.accounts {
			if err := e.syncAccount(ctx, account); err != nil {
				return err
			}
		}
		return nil
	})
}

func (e MetaSyncer[T]) syncAccount(ctx context.Context, acc MetaAccount[T]) error {
	things := acc.MetaThings()

	for _, thing := range things {
		// Get setting.
		setting := thing.SyncSetting()
		if !setting.Synchronize() {
			return nil
		}

		// Sync.
		theSyncer := syncerimpl.NewMetaAccount(
			e.repo,
			thing.EntityID(),
			setting.LastSynchronization(),
			setting.SetLastSynchronization,
			thing.Param(),
			thing.SetParam,
		)
		if err := theSyncer.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}
