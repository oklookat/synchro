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
		Info("lovesYou", "linker.Static", "~~~ WISH ME LUCK! <3 ~~~")
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
func (e Static) FromRemote(ctx context.Context, target RemoteEntity) (FromRemoteResult, error) {
	slog.Info("name", target.Name(), "remoteID", target.ID().String(), "from", target.RemoteName().String())

	result := FromRemoteResult{}
	result.RemoteEntity = target

	targetRemote, ok := e.remotes[target.RemoteName()]
	if !ok {
		return result, shared.NewErrRemoteNotFound(target.RemoteName())
	}

	// Link exists?
	linked, err := targetRemote.Linkables().LinkedRemoteID(target.ID())
	if err != nil {
		return result, err
	}

	// Link exists.
	if !shared.IsNil(linked) {
		// Missing before?
		if linked.RemoteID() == nil {
			// Set ID.
			slog.Info("SET ID (MISSING BEFORE)")
			updId := target.ID()
			if err = linked.SetRemoteID(&updId); err != nil {
				return result, err
			}
		}
		result.Linked = linked
		return result, err
	}

	// Link not exits.

	// Find an entity to link with target.
	entitiesResult, err := e.findEntitiesToLink(ctx, target)
	if err != nil {
		return result, err
	}

	// Not linked by all remotes.
	if entitiesResult.FoundEntity == nil {
		// Create new entity.
		entityId, err := e.repo.CreateEntity()
		if err != nil {
			return result, err
		}
		entitiesResult.FoundEntity = &entityId
	}
	entityId := *entitiesResult.FoundEntity

	// Link entity with remotes.
	for remName, remId := range entitiesResult.TargetRemoteId {
		// Get remote.
		remote := e.remotes[remName]

		// Link exists?
		linked, err := remote.Linkables().LinkedEntity(entityId)
		if err != nil {
			return result, err
		}

		// Exists.
		if !shared.IsNil(linked) {
			// Change ID?
			//
			// Example 1: an artist deleted his Spotify profile,
			// but we have his ID in the database. So the data in the database is out of date,
			// and we need to mark the artist as missing on Spotify.
			//
			// Example 2: for some reason the artist ID has changed.
			// For example, the artist has 2 Spotify profiles - an old and a new one.
			// And for some reason the matcher chose the old one instead of the new one.
			// This means that either the artist deleted his new profile
			// or the matcher made a mistake(?). I call it "relinking".
			// It can be not only because of the example above, but also if you
			// deliberately liked the old profile instead of the new one.
			// Or because of different catalogs on different streaming services.
			// I can do something about it, with several links to one ID,
			// and so on, but it's a waste of time. The artist should have one profile.
			// Writing a bunch of code because of errors and defects of streaming services is bullshit. KISS.

			isIdNotChanged := ((linked.RemoteID() != nil && remId != nil) &&
				(*linked.RemoteID() == *remId))
			if !isIdNotChanged {
				if err := linked.SetRemoteID(remId); err != nil {
					return result, err
				}
			}
			continue
		} else {
			newLinked, err := remote.Linkables().CreateLink(ctx, entityId, remId)
			if err != nil {
				return result, err
			}
			linked = newLinked
		}

		// Add to result if current remote is a target.
		if remName == target.RemoteName() {
			result.Linked = linked
		}
	}

	return result, err
}

// From entity to remote entity.
func (e Static) ToRemote(ctx context.Context, id shared.EntityID, target shared.RemoteName) (ToRemoteResult, error) {
	result := ToRemoteResult{}

	targetRem, ok := e.remotes[target]
	if !ok {
		return result, shared.NewErrRemoteNotFound(target)
	}

	defer e.repo.DeleteNotLinked()

	// Link exists?
	linked, err := targetRem.Linkables().LinkedEntity(id)
	if err != nil {
		return result, err
	}

	// Exists.
	if !shared.IsNil(linked) {
		result.Linked = linked

		// Missing?
		result.MissingBefore = linked.RemoteID() == nil
		result.MissingNow = result.MissingBefore
		if !result.MissingBefore {
			// Not missing.
			return result, err
		}

		// Recheck? Maybe it's not missing now.

		cfg, err := config.Get[config.Linker](config.KeyLinker)
		if err != nil {
			return result, err
		}

		if !cfg.RecheckMissing {
			// Recheck disabled in config.
			return result, err
		}
	}

	// Entity not linked with target or missing before.
	// Create/find link.

	// Find link in another remotes.
	for name := range e.remotes {
		// Skip current.
		if name == target {
			continue
		}

		// Link exists?
		linkedFromAnother, err := e.remotes[name].Linkables().LinkedEntity(id)
		if err != nil {
			return result, err
		}

		// Not exists?
		if shared.IsNil(linkedFromAnother) {
			// Try another.
			continue
		}

		// Exists, but missing.
		if linkedFromAnother.RemoteID() == nil {
			// Skip.
			continue
		}

		// Exists, and not missing. Try to get entity in remote.
		entityFromAnotherRemote, err := e.remotes[name].RemoteEntity(ctx, *linkedFromAnother.RemoteID())
		if err != nil {
			return result, err
		}

		// Not exists?
		if shared.IsNil(entityFromAnotherRemote) {
			// Make missing.
			if err = linkedFromAnother.SetRemoteID(nil); err != nil {
				return result, err
			}
			// Try another.
			continue
		}

		// Exists. Search entity from another remote in target.
		found, err := e.search(ctx, entityFromAnotherRemote, target)
		if err != nil {
			return result, err
		}

		// Not found?
		if shared.IsNil(found) {
			// Make/stay missing.
			result.MissingNow = true

			// Link exists before?
			if !shared.IsNil(result.Linked) {
				// Stay missing.
				return result, err
			}

			// Link not exists.
			// Create link, mark as missing.
			result.NewLink = true
			linked, err := targetRem.Linkables().CreateLink(ctx, id, nil)
			result.Linked = linked
			return result, err
		}

		// Found. Missing before, but now not.
		foundID := found.ID()
		result.MissingNow = false

		// Link exists?
		if !shared.IsNil(result.Linked) {
			// Set ID.
			return result, linked.SetRemoteID(&foundID)
		}

		// Link not exists. Create.
		result.NewLink = true
		linked, err := targetRem.Linkables().CreateLink(ctx, id, &foundID)
		if err != nil {
			return result, err
		}

		result.Linked = linked
		return result, err
	}

	return result, errors.New("broken links")
}

type findEntitiesToLinkResult struct {
	// Target used for fine entities to link.
	Target RemoteEntity

	// Entity that can be linked with Target.
	FoundEntity *shared.EntityID

	// 1. Target ID in another remotes that also not linked with FoundEntities.
	// Includes Target.
	//
	// 2. Target ID in another remotes that missing in remote.
	TargetRemoteId map[shared.RemoteName]*shared.RemoteID
}

// Find an entities to link with target.
func (e Static) findEntitiesToLink(ctx context.Context, target RemoteEntity) (findEntitiesToLinkResult, error) {
	result := findEntitiesToLinkResult{
		Target:         target,
		TargetRemoteId: map[shared.RemoteName]*shared.RemoteID{},
	}

	slog.Info("FIND ENTITY FOR", "from", target.RemoteName().String(), "name", target.Name(), "remoteID", target.ID().String())

	// Add current.
	targetId := target.ID()
	result.TargetRemoteId[target.RemoteName()] = &targetId

	for _, remote := range e.remotes {
		// Skip current.
		if remote.Name() == target.RemoteName() {
			continue
		}

		// Find target.
		foundTarget, err := e.search(ctx, target, remote.Name())
		if err != nil {
			return result, err
		}

		// Missing?
		if shared.IsNil(foundTarget) {
			// Mark as missing.
			result.TargetRemoteId[remote.Name()] = nil
			continue
		}

		// Linked?
		linked, err := remote.Linkables().LinkedRemoteID(foundTarget.ID())
		if err != nil {
			return result, err
		}

		// Not linked.
		if shared.IsNil(linked) {
			foundID := foundTarget.ID()
			result.TargetRemoteId[remote.Name()] = &foundID
			continue
		}

		// Linked. Get entity id.
		entityId := linked.EntityID()
		result.FoundEntity = &entityId
		return result, err
	}

	return result, nil
}

// Search any remote entity in any remote.
//
// Example: search Spotify artist in Yandex.Music.
//
// Returns remote entity from target.
//
// Nil if not exists.
func (e Static) search(ctx context.Context, who RemoteEntity, target shared.RemoteName) (RemoteEntity, error) {
	targetRem, ok := e.remotes[target]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(target)
	}

	slog.Info("==== 🔎 ====", "from", who.RemoteName().String(),
		"name", who.Name(),
		"remoteID", who.ID().String(),
		"to", target.String())

	// same remotes.
	if target == who.RemoteName() {
		slog.Info("✅ (same remotes)")
		return who, nil
	}

	// Match.
	matched, err := targetRem.Match(ctx, who)
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
