package syncer

import (
	"context"
	"time"

	"github.com/oklookat/synchro/shared"
)

type (
	// Example: remote-local entity.
	SyncableEntity[T comparable] interface {
		// Remote entity param.
		//
		// Example: is_removed, playlist name, etc.
		SyncableParam[T]
	}

	// Database.
	Repository[T comparable] interface {
		// Create synced param that linked with entity.
		//
		// SyncParam time must be 0.
		CreateSynced(context.Context, shared.EntityID) (Synced[T], error)

		// Get synced entity.
		Synced(shared.EntityID) (Synced[T], error)

		// Get Synced newer than date.
		SyncesNewer(ctx context.Context, newerThan time.Time) (map[shared.EntityID]Synced[T], error)

		// Get Synced older than date.
		SyncesOlder(ctx context.Context, olderThan time.Time) (map[shared.EntityID]Synced[T], error)

		DeleteUnsynced() error
	}

	// Synced param in DB.
	Synced[T comparable] interface {
		SyncableParam() SyncableParam[T]

		// Must be updated to current time after calling Set().
		ModifiedAt() time.Time
	}

	// Examples:
	//
	// 1. Liked: Set() = unlikes track (or adds track to unlike queue, depends on implementation).
	// Get() = in case of remote - returns true (track in user library).
	// In case of DB returns bool depending on "is_removed" field or something like this.
	//
	// 2. Playlist description: Set() = set description. Get() - get description.
	SyncableParam[T comparable] interface {
		Get() T
		Set(context.Context, T) error
	}
)

func New[T comparable](repo Repository[T], remoteData map[shared.EntityID]SyncableEntity[T], remoteLastSync time.Time) *Instance[T] {
	return &Instance[T]{
		repo:           repo,
		remoteData:     remoteData,
		remoteLastSync: remoteLastSync,
	}
}

// Usage:
//
// 1. First, call FromRemote.
//
// 2. Then ToRemote (if you working with static entities).
type Instance[T comparable] struct {
	repo Repository[T]

	// Example: account liked artists.
	remoteData map[shared.EntityID]SyncableEntity[T]

	// Example: account liked artists last sync.
	remoteLastSync time.Time
}

// Remote > DB.
func (e Instance[T]) FromRemote(ctx context.Context) error {
	// Check Remote data changes.
	for id := range e.remoteData {
		syncd, err := e.repo.Synced(id)
		if err != nil {
			return err
		}

		// Not exists in DB?
		if shared.IsNil(syncd) {
			// Create.
			syncd, err = e.repo.CreateSynced(ctx, id)
			if err != nil {
				return err
			}

			// Set.
			if err = syncd.SyncableParam().Set(ctx, e.remoteData[id].Get()); err != nil {
				return err
			}
			continue
		}

		// Compare and set data if needed.
		if err = e.setCompareSyncableSynced(ctx, e.remoteData[id], syncd); err != nil {
			return err
		}
	}

	return nil
}

// # Useful for static (and possibly missing) entities
//
// * static entities examples: tracks, artists, albums.
//
// * Get, Set methods below used on Synced.SyncableParam.
//
// # newd
//
// That entities not exists in Remote data,
// and have newer sync date than account last sync.
//
// 2 options:
//
// 1. Get() == true, you need to like it in Remote. If false - do nothing.
//
// 2. Entity missing in Remote. Do nothing.
//
// # old
//
// That entities not exists in remote data,
// and account sync date is newer than the entities sync date.
//
// 4 options:
//
// 1. Get() == false.
// Need to check is entity will be missing before in Remote.
// If yes, like it on remote, and Set(true).
// Otherwise go to p.2.
//
// 2. Get() == false.
// Entity is not liked. Do nothing.
//
// 3. Get() == true. Check is entity will be missing before in Remote.
// If yes, need to like it in Remote.
// Otherwise go to p.4.
//
// 4. Get() == true. Set(false), because the DB has outdated data
// (account sync date newer than DB sync date).
func (e Instance[T]) ToRemote(ctx context.Context) (newd map[shared.EntityID]Synced[T], old map[shared.EntityID]Synced[T], err error) {
	newd = map[shared.EntityID]Synced[T]{}
	old = map[shared.EntityID]Synced[T]{}

	// Add / change.
	synces, err := e.repo.SyncesNewer(ctx, e.remoteLastSync)
	if err != nil {
		return
	}

	for id := range synces {
		rem, ok := e.remoteData[id]
		if !ok {
			newd[id] = synces[id]
			continue
		}
		if err = e.setCompareSyncableSynced(ctx, rem, synces[id]); err != nil {
			return
		}
	}

	// Remove / not present in Remote data.
	// Check data that not present in remote.
	synces, err = e.repo.SyncesOlder(ctx, e.remoteLastSync)
	if err != nil {
		return
	}

	for id := range synces {
		if len(e.remoteData) > 0 {
			_, ok := e.remoteData[id]
			if ok {
				continue
			}
		}
		old[id] = synces[id]
	}

	return
}

// Compare and set data to entity or synced if changes maded (by sync date comparing).
func (e Instance[T]) setCompareSyncableSynced(ctx context.Context, entity SyncableEntity[T], synced Synced[T]) error {
	// Value not changed.
	if entity.Get() == synced.SyncableParam().Get() {
		return nil
	}

	// Remote data newer.
	//
	// Or same sync time. Remote priority
	if e.remoteLastSync.After(synced.ModifiedAt()) || e.remoteLastSync.Equal(synced.ModifiedAt()) {
		// Example: unlike artist in DB.
		if err := synced.SyncableParam().Set(ctx, entity.Get()); err != nil {
			return err
		}
		return nil
	}

	// DB data newer.
	// Example: unlike artist in Remote.
	return entity.Set(ctx, synced.SyncableParam().Get())
}
