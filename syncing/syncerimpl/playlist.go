package syncerimpl

import (
	"context"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

type (
	PlaylistDeleter interface {
		// Delete my playlists.
		Delete(ctx context.Context, entities []shared.RemoteID) error
	}
)

func NewPlaylistAccount(
	repo syncer.Repository[bool],
	lnk *linker.Dynamic,

	lastSync time.Time,
	setLastSync func(time.Time) error,
	accountID string,
	playlists map[shared.EntityID]shared.RemoteID,
	deleter PlaylistDeleter,
) *PlaylistAccount {
	if repo == nil || lnk == nil ||
		len(accountID) == 0 || setLastSync == nil ||
		deleter == nil {
		return nil
	}
	return &PlaylistAccount{
		repo:        repo,
		lnk:         lnk,
		lastSync:    lastSync,
		setLastSync: setLastSync,
		AccountID:   accountID,
		playlists:   playlists,
		deleter:     deleter,
		NewRemoved:  map[shared.EntityID]shared.RemoteID{},
		RemoveLinks: map[shared.EntityID]bool{},
	}
}

type PlaylistAccount struct {
	repo syncer.Repository[bool]
	lnk  *linker.Dynamic

	// Example: playlists last sync.
	lastSync time.Time

	// Example: set playlists last sync.
	setLastSync func(time.Time) error

	AccountID string
	playlists map[shared.EntityID]shared.RemoteID
	deleter   PlaylistDeleter

	// New things that removed.
	NewRemoved map[shared.EntityID]shared.RemoteID

	// Links that will be deleted (because playlists removed).
	RemoveLinks map[shared.EntityID]bool
}

func (e *PlaylistAccount) Start(ctx context.Context) error {
	remoteData := map[shared.EntityID]syncer.SyncableEntity[bool]{}

	for id := range e.playlists {
		remoteData[id] = &likeableSyncableParam{
			id:  id,
			get: true,

			likes:   e.playlists,
			unliked: e.NewRemoved,
		}
	}

	sync := syncer.New(e.repo, remoteData, e.lastSync)
	if err := sync.FromRemote(ctx); err != nil {
		return err
	}

	_, notPresentOld, err := sync.ToRemote(ctx)
	if err != nil {
		return err
	}

	// Not present in this case: account sync newer and playlists deleted from account.
	for eId := range notPresentOld {
		// Mark as deleted in DB.
		if err := notPresentOld[eId].SyncableParam().Set(ctx, false); err != nil {
			return err
		}
	}

	if len(e.NewRemoved) > 0 {
		if err := e.delete(ctx, e.NewRemoved); err != nil {
			return err
		}
	}

	return e.setLastSync(time.Now())
}

func (e *PlaylistAccount) delete(ctx context.Context, data map[shared.EntityID]shared.RemoteID) error {
	var toDelete []shared.RemoteID

	for eId := range data {
		e.RemoveLinks[eId] = true
		_, ok := e.playlists[eId]
		if !ok {
			// Not exits in user library.
			delete(e.playlists, eId)
			continue
		}

		linked, err := e.lnk.ToRemote(ctx, eId, shared.RemoteName(e.AccountID), false)
		if err != nil {
			return err
		}

		if shared.IsNil(linked) {
			delete(e.playlists, eId)
			continue
		}

		toDelete = append(toDelete, linked.RemoteID())
	}

	if len(toDelete) == 0 {
		return nil
	}

	if err := e.deleter.Delete(ctx, toDelete); err == nil {
		for eId := range data {
			delete(e.playlists, eId)
		}
	}

	return nil
}
