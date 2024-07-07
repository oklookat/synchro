package linker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
)

type (
	// Links global entities like artists, tracks, albums.
	Remote interface {
		// Unique name.
		//
		// Example: "Spotify".
		Name() shared.RemoteName

		// Find entity from another remote in current.
		Match(context.Context, RemoteEntity) (RemoteEntity, error)

		// Get entity by ID.
		RemoteEntity(context.Context, shared.RemoteID) (RemoteEntity, error)

		// DB ops for linked entities.
		Linkables() Linkables
	}

	// Remote in DB.
	Linkables interface {
		// Link entity with remote.
		CreateLink(context.Context, shared.EntityID, *shared.RemoteID) (Linked, error)

		// Example: get linked spotify artist by artist entity.
		LinkedEntity(shared.EntityID) (Linked, error)

		// Example: get linked spotify artist by its ID on Spotify.
		LinkedRemoteID(shared.RemoteID) (Linked, error)
	}

	RemoteEntity interface {
		// Example: Spotify.
		RemoteName() shared.RemoteName

		// Example: Spotify artist ID.
		ID() shared.RemoteID

		// Example: Spotify artist name.
		Name() string
	}

	Linked interface {
		// Parent entity.
		EntityID() shared.EntityID

		// Example: Spotify artist ID.
		//
		// Nil if not exists in remote.
		RemoteID() *shared.RemoteID

		// Example: set Spotify artist ID.
		//
		// Nil if not exists in remote.
		SetRemoteID(*shared.RemoteID) error

		// Date when link created/modified.
		ModifiedAt() time.Time
	}

	ToRemoteResult struct {
		// Missing on remote before but now present?
		MissingBefore,

		// Missing now.
		MissingNow,

		// Link not exists before?
		NewLink bool

		// Entity linked with target remote.
		//
		// Can be nil.
		Linked Linked
	}

	FromRemoteResult struct {
		// Remote entity from target.
		RemoteEntity RemoteEntity

		// Linked RemoteEntity.
		Linked Linked
	}
)

func NewStatic(repo Repository, remotes map[shared.RemoteName]Remote) *Static {
	slog.
		Info("lovesYou", "linker (static)", "~~~ WISH ME LUCK! <3 ~~~")
	return &Static{
		repo:    repo,
		remotes: remotes,
	}
}

// Links global entities.
//
// Example: track, artist, album.
type Static struct {
	repo    Repository
	remotes map[shared.RemoteName]Remote
}

// From remote entity to linked.
func (e Static) FromRemote(ctx context.Context, source RemoteEntity, target shared.RemoteName) (FromRemoteResult, error) {
	slog.Info("fromRemote", "name", source.Name(), "remoteID", source.ID().String(), "from", source.RemoteName().String())

	result := FromRemoteResult{}
	result.RemoteEntity = source

	sourceRemote, ok := e.remotes[source.RemoteName()]
	if !ok {
		return result, shared.NewErrRemoteNotFound(source.RemoteName())
	}

	// Link exists?
	sourceLinked, err := sourceRemote.Linkables().LinkedRemoteID(source.ID())
	if err != nil {
		return result, err
	}

	// Link exists.
	if !shared.IsNil(sourceLinked) {
		// Missing before?
		if sourceLinked.RemoteID() == nil {
			// Set ID.
			slog.Info("SET ID (MISSING BEFORE)")
			updId := source.ID()
			if err = sourceLinked.SetRemoteID(&updId); err != nil {
				return result, err
			}
		}
		result.Linked = sourceLinked
		return result, err
	}

	// Link not exits.

	targetRemote, ok := e.remotes[target]
	if !ok {
		return result, shared.NewErrRemoteNotFound(target)
	}

	// Find an entity to link with target.
	found, foundLinked, err := e.findEntityToLink(ctx, source, target)
	if err != nil {
		return result, err
	}

	// Found id?
	var foundIdTarget *shared.RemoteID
	if !shared.IsNil(found) {
		gg := found.ID()
		foundIdTarget = &gg
	}

	var entityLinkTo shared.EntityID

	// Target not linked?
	if shared.IsNil(foundLinked) {
		// Create new entity.
		entityLinkTo, err = e.repo.CreateEntity()
		if err != nil {
			return result, err
		}
		// Link with target.
		_, err = targetRemote.Linkables().CreateLink(context.Background(), entityLinkTo, foundIdTarget)
		if err != nil {
			return result, err
		}
	} else {
		// Target linked.
		entityLinkTo = foundLinked.EntityID()
		// Change id?
		isIdNotChanged := ((foundLinked.RemoteID() != nil && foundIdTarget != nil) &&
			(*foundLinked.RemoteID() == *foundIdTarget))
		if !isIdNotChanged {
			if err := foundLinked.SetRemoteID(foundIdTarget); err != nil {
				return result, err
			}
		}
	}

	// Link with source.
	srcId := source.ID()
	sourceLinked, err = sourceRemote.Linkables().CreateLink(context.Background(), entityLinkTo, &srcId)
	if err != nil {
		return result, err
	}

	result.Linked = sourceLinked

	return result, err
}

// From source linked entity to target linked entity.
func (e Static) ToRemote(ctx context.Context, sourceLinked Linked, source, target shared.RemoteName) (ToRemoteResult, error) {
	result := ToRemoteResult{}

	// No remote id?
	if sourceLinked.RemoteID() == nil {
		// WTF?
		return result, errors.New("broken links")
	}

	// Get remotes.
	sourceRem, ok := e.remotes[source]
	if !ok {
		return result, shared.NewErrRemoteNotFound(target)
	}
	targetRem, ok := e.remotes[target]
	if !ok {
		return result, shared.NewErrRemoteNotFound(target)
	}

	// Delete strange links.
	defer e.repo.DeleteNotLinked()

	// Source linked with target?
	targetLinked, err := targetRem.Linkables().LinkedEntity(sourceLinked.EntityID())
	if err != nil {
		return result, err
	}

	linkedWithTarget := !shared.IsNil(targetLinked)

	// Linked.
	if linkedWithTarget {
		result.Linked = targetLinked

		// Missing?
		result.MissingBefore = targetLinked.RemoteID() == nil
		result.MissingNow = result.MissingBefore
		if !result.MissingBefore {
			// Not missing.
			return result, err
		}

		// Recheck? Maybe it's not missing now.

		cfg, err := config.Get[*config.Linker](config.KeyLinker)
		if err != nil {
			return result, err
		}

		if !(*cfg).RecheckMissing {
			// Recheck disabled in config.
			return result, err
		}

	}

	// Not linked with target OR linked, but missing (need to recheck).

	// Try to get entity from source remote.
	entityFromSourceRemote, err := sourceRem.RemoteEntity(ctx, *sourceLinked.RemoteID())
	if err != nil {
		return result, err
	}

	// Not exists?
	if shared.IsNil(entityFromSourceRemote) {
		result.MissingNow = true
		// Probably entity deleted from remote. Mark both as missing.
		if err = sourceLinked.SetRemoteID(nil); err != nil {
			return result, err
		}
		if linkedWithTarget {
			if err := targetLinked.SetRemoteID(nil); err != nil {
				return result, err
			}
			return result, err
		}
		result.NewLink = true
		linked, err := targetRem.Linkables().CreateLink(ctx, sourceLinked.EntityID(), nil)
		if err != nil {
			return result, err
		}
		result.Linked = linked
		return result, err
	}

	// Exists in source. Search entity from source remote in target.
	foundInTarget, err := e.search(ctx, entityFromSourceRemote, target)
	if err != nil {
		return result, err
	}

	// Not found in target remote?
	if shared.IsNil(foundInTarget) {
		if linkedWithTarget {
			// Stay missing.
			return result, err
		}
		// Create link, mark as missing.
		result.NewLink = true
		linked, err := targetRem.Linkables().CreateLink(ctx, sourceLinked.EntityID(), nil)
		result.Linked = linked
		return result, err
	}

	// Found.
	foundID := foundInTarget.ID()
	result.MissingNow = false

	// Link exists?
	if linkedWithTarget {
		return result, targetLinked.SetRemoteID(&foundID)
	}

	// Link not exists. Create.
	result.NewLink = true
	targetLinked, err = targetRem.Linkables().CreateLink(ctx, sourceLinked.EntityID(), &foundID)
	if err != nil {
		return result, err
	}

	result.Linked = targetLinked
	return result, err
}

// Find an entities to link with target.
//
// Returns:
//
// 1. Found SOURCE in TARGET, linked TARGET.
//
// 2. nil, nil, nil if SOURCE not found in TARGET.
//
// 3. ok, nil, nil if SOURCE found in TARGET, but not linked.
func (e Static) findEntityToLink(ctx context.Context, source RemoteEntity, target shared.RemoteName) (RemoteEntity, Linked, error) {
	slog.Info("FIND ENTITY FOR", "from", source.RemoteName().String(), "name", source.Name(), "remoteID", source.ID().String())

	// Find target.
	foundTarget, err := e.search(ctx, source, target)
	if err != nil {
		return nil, nil, err
	}

	// Missing?
	if shared.IsNil(foundTarget) {
		// Missing.
		return nil, nil, nil
	}

	// Linked?
	targetRem, ok := e.remotes[target]
	if !ok {
		return nil, nil, shared.NewErrRemoteNotFound(target)
	}
	linked, err := targetRem.Linkables().LinkedRemoteID(foundTarget.ID())
	if err != nil {
		return nil, nil, err
	}

	// Not linked.
	if shared.IsNil(linked) {
		return foundTarget, nil, err
	}

	// All found.
	return foundTarget, linked, err
}

// Search any remote entity in any remote.
//
// Example: search Spotify artist in Yandex.Music.
//
// Returns remote entity from target.
//
// Nil if not exists.
func (e Static) search(ctx context.Context, source RemoteEntity, target shared.RemoteName) (RemoteEntity, error) {
	targetRem, ok := e.remotes[target]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(target)
	}

	slog.Info("==== 🔎 ====", "from", source.RemoteName().String(),
		"name", source.Name(),
		"remoteID", source.ID().String(),
		"to", target.String())

	// same remotes.
	if target == source.RemoteName() {
		slog.Info("✅ (same remotes)")
		return source, nil
	}

	// Match.
	matched, err := targetRem.Match(ctx, source)
	if err != nil {
		return nil, err
	}
	if shared.IsNil(matched) {
		slog.Info("❌")
		return nil, err
	}

	slog.Info("✅", "matchedRemoteID", matched.ID().String())

	return matched, err
}
