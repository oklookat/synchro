package syncerimpl

import (
	"context"
	"log/slog"
	"time"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

type LikeUnliker interface {
	// Examples: like artists.
	Like(ctx context.Context, ids []shared.RemoteID) error

	// Examples: unlike artists.
	Unlike(ctx context.Context, ids []shared.RemoteID) error
}

// Example: liked artists syncer per account.
func NewLikeableAccount(
	repo syncer.Repository[bool],
	lnk *linker.Static,

	remoteName shared.RemoteName,
	likes map[shared.EntityID]shared.RemoteID,
	lastSync time.Time,
	setLastSync func(time.Time) error,
	actions LikeUnliker,
) *LikeableAccount {
	if repo == nil || lnk == nil ||
		len(remoteName) == 0 || setLastSync == nil ||
		actions == nil {
		return nil
	}
	return &LikeableAccount{
		Repo:        repo,
		lnk:         lnk,
		remoteName:  remoteName,
		likes:       likes,
		lastSync:    lastSync,
		setLastSync: setLastSync,
		actions:     actions,

		NewUnliked: map[shared.EntityID]shared.RemoteID{},
		NewLiked:   map[shared.EntityID]shared.RemoteID{},
	}
}

type LikeableAccount struct {
	// Example: artists repo.
	Repo syncer.Repository[bool]

	// Example: artists linker.
	lnk *linker.Static

	// Example: Spotify.
	remoteName shared.RemoteName

	// Example: liked artists.
	likes map[shared.EntityID]shared.RemoteID

	// Example: liked artists last sync.
	lastSync time.Time

	// Example: set liked artists last sync.
	setLastSync func(time.Time) error

	// Example: liked artists actions.
	actions LikeUnliker

	// Unliked some tracks?
	IsChangesUnliked bool

	// New things that unliked.
	NewUnliked map[shared.EntityID]shared.RemoteID

	// Liked some tracks?
	IsChangesLiked bool

	// New things that liked.
	NewLiked map[shared.EntityID]shared.RemoteID
}

func (e *LikeableAccount) Start(ctx context.Context) error {
	remoteData := map[shared.EntityID]syncer.SyncableEntity[bool]{}

	for id := range e.likes {
		remoteData[id] = &likeableSyncableParam{
			id:  id,
			get: true,

			likes:   e.likes,
			unliked: e.NewUnliked,
		}
	}

	sync := syncer.New(e.Repo, remoteData, e.lastSync)
	if err := sync.FromRemote(ctx); err != nil {
		return err
	}
	notPresentNew, notPresentOld, err := sync.ToRemote(ctx)
	if err != nil {
		return err
	}

	toLikeNotPresent, err := e.checkNotPresent(ctx, notPresentOld)
	if err != nil {
		return err
	}
	totalToLike := map[shared.EntityID]syncer.Synced[bool]{}
	for id, v := range toLikeNotPresent {
		totalToLike[id] = v
	}
	for id, v := range notPresentNew {
		totalToLike[id] = v
	}

	var toLike []shared.EntityID
	for id := range totalToLike {
		_, ok := notPresentNew[id]
		if ok {
			// Not liked in DB.
			if !notPresentNew[id].SyncableParam().Get() {
				continue
			}
		}

		// Already liked in remote
		// (syncer / DB mistake?).
		if _, ok := e.likes[id]; ok {
			slog.Warn("not present, but already liked", "entityID", id)
			continue
		}
		toLike = append(toLike, id)
	}

	if err := e.like(ctx, toLike); err != nil {
		e.autoRecover()
		return err
	}
	if err := e.unlike(ctx, e.NewUnliked); err != nil {
		e.autoRecover()
		return err
	}

	return e.setLastSync(time.Now())
}

func (e *LikeableAccount) autoRecover() error {
	ctx := context.Background()

	snapCfg, err := config.Get[config.Snapshots](config.KeySnapshots)
	if err != nil {
		return err
	}
	if !snapCfg.AutoRecover {
		return nil
	}
	if e.IsChangesLiked {
		slog.Info("auto recover (unlike liked)")
		var ids shared.RemoteIDSlice[shared.EntityID]
		ids.FromMapV(e.NewLiked)
		if err := e.actions.Unlike(ctx, ids); err != nil {
			return err
		}
		e.IsChangesLiked = false
	}
	if e.IsChangesUnliked {
		slog.Info("auto recover (like unliked)")
		var ids shared.RemoteIDSlice[shared.EntityID]
		ids.FromMapV(e.NewUnliked)
		if err := e.actions.Like(ctx, ids); err != nil {
			return err
		}
		e.IsChangesUnliked = false
	}
	return nil
}

func (e *LikeableAccount) like(ctx context.Context, data []shared.EntityID) error {
	if len(data) == 0 {
		return nil
	}

	converted := map[shared.EntityID]shared.RemoteID{}
	var ids []shared.RemoteID

	for _, id := range data {
		result, err := e.lnk.ToRemote(ctx, id, e.remoteName)
		if err != nil {
			return err
		}
		if shared.IsNil(result.Linked) || result.Linked.RemoteID() == nil {
			continue
		}
		converted[id] = *result.Linked.RemoteID()
		ids = append(ids, converted[id])
	}

	if len(converted) == 0 {
		return nil
	}

	if err := e.actions.Like(ctx, ids); err == nil {
		e.IsChangesLiked = true
		for eId, rId := range converted {
			e.likes[eId] = rId
		}
	}

	return nil
}

func (e *LikeableAccount) unlike(ctx context.Context, data map[shared.EntityID]shared.RemoteID) error {
	if len(data) == 0 {
		return nil
	}

	var ids []shared.RemoteID
	for id := range data {
		remoteID, ok := e.likes[id]
		if !ok {
			continue
		}
		ids = append(ids, remoteID)
	}

	if err := e.actions.Unlike(ctx, ids); err == nil {
		e.IsChangesUnliked = true
		for id := range data {
			delete(e.likes, id)
		}
	}

	return nil
}

func (e *LikeableAccount) checkNotPresent(
	ctx context.Context,
	data map[shared.EntityID]syncer.Synced[bool],
) (map[shared.EntityID]syncer.Synced[bool], error) {

	result := map[shared.EntityID]syncer.Synced[bool]{}

	for id := range data {
		linkerResult, err := e.lnk.ToRemote(ctx, shared.EntityID(id), e.remoteName)
		if err != nil {
			return nil, err
		}

		if linkerResult.MissingNow {
			continue
		}

		// Missing before?
		if linkerResult.MissingBefore {
			// Marked in DB as not liked?
			if !data[id].SyncableParam().Get() {
				// Make liked in DB.
				if err = data[id].SyncableParam().Set(ctx, true); err != nil {
					return nil, err
				}
			}

			// Add to remote account likes.
			result[id] = data[id]
			continue
		}

		// Not missing before.

		// Marked in DB as not liked?
		if !data[id].SyncableParam().Get() {
			continue
		}

		// Mark as not liked, because current account lastSync is newer than DB.
		if err = data[id].SyncableParam().Set(ctx, false); err != nil {
			return nil, err
		}
	}

	return result, nil
}
