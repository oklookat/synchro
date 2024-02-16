package linker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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

		// All links.
		Links() ([]Linked, error)

		// Count of not matched entities.
		NotMatchedCount() (int, error)
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

	// Links import/export database.
	IEData struct {
		Remotes map[shared.RemoteName]struct {
			Entities map[shared.EntityID]*shared.RemoteID `json:"entities"`
		} `json:"remotes"`
	}
)

func NewStatic(repo Repository, remotes map[shared.RemoteName]Remote) *Static {
	_log.AddField("lovesYou", "linker.Static").
		Info("~~~ WISH ME LUCK! <3 ~~~")
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
	theLog := _log.
		AddField("name", target.Name()).
		AddField("remoteID", target.ID().String()).
		AddField("from", target.RemoteName().String())

	result := FromRemoteResult{}
	result.RemoteEntity = target

	targetRemote, ok := e.remotes[target.RemoteName()]
	if !ok {
		return result, shared.NewErrRemoteNotFound(_packageName, target.RemoteName())
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
			theLog.Info("SET ID (MISSING BEFORE)")
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
		return result, shared.NewErrRemoteNotFound(_packageName, target)
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

		cfg := &config.Linker{}
		if err := config.Get(cfg); err != nil {
			cfg.Default()
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

// Import links for all remotes.
func (e Static) Import(r io.Reader) error {
	data := &IEData{}
	if err := json.NewDecoder(r).Decode(data); err != nil {
		return err
	}

	defer e.repo.DeleteNotLinked()

	skip := map[shared.EntityID]bool{}

	// Step 1: skip links that already exists.
	for remoteName, remote := range e.remotes {
		dataRemote, ok := data.Remotes[remoteName]

		// Skip unknown remote.
		if !ok {
			continue
		}

		// Find any remote that have link.
		for eId, remId := range dataRemote.Entities {
			if _, ok := skip[eId]; ok {
				continue
			}

			// Exists?
			link, err := remote.Linkables().LinkedRemoteID(*remId)
			if err != nil {
				return err
			}

			// Exists.
			if !shared.IsNil(link) {
				// Skip.
				skip[eId] = true
				break
			}
		}
	}

	// Step 2: create links.

	// [ID FROM DATA]ID FROM CURRENT DB.
	dataDbMapping := map[shared.EntityID]shared.EntityID{}

	for remoteName, remote := range e.remotes {
		dataRemote, ok := data.Remotes[remoteName]
		if !ok {
			// Skip unknown remote.
			continue
		}
		for eId, remId := range dataRemote.Entities {
			// Skip already linked.
			if _, ok := skip[eId]; ok {
				continue
			}

			var entityToLink shared.EntityID
			if entID, ok := dataDbMapping[eId]; ok {
				// Entity created before.
				entityToLink = entID
			} else {
				// New entity.
				newEntity, err := e.repo.CreateEntity()
				if err != nil {
					return err
				}
				entityToLink = newEntity
				// Mark as *Entity created before*.
				dataDbMapping[eId] = newEntity
			}

			// Make link.
			var copyRemId *shared.RemoteID
			if remId != nil {
				realCopy := *remId
				copyRemId = &realCopy
			}
			_, err := remote.Linkables().CreateLink(context.Background(), entityToLink, copyRemId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Export links from all remotes to json.
func (e Static) Export(dest io.Writer) error {
	result := &IEData{}
	result.Remotes = map[shared.RemoteName]struct {
		Entities map[shared.EntityID]*shared.RemoteID "json:\"entities\""
	}{}

	for remoteName, remote := range e.remotes {
		result.Remotes[remoteName] = struct {
			Entities map[shared.EntityID]*shared.RemoteID "json:\"entities\""
		}{}
		resRem := result.Remotes[remoteName]
		resRem.Entities = map[shared.EntityID]*shared.RemoteID{}
		result.Remotes[remoteName] = resRem

		links, err := remote.Linkables().Links()
		if err != nil {
			return err
		}

		for _, linked := range links {
			var remId *shared.RemoteID
			if linked.RemoteID() != nil {
				copyId := *linked.RemoteID()
				remId = &copyId
			}
			result.Remotes[remoteName].Entities[linked.EntityID()] = remId
		}
	}

	return json.NewEncoder(dest).Encode(result)
}

// Clean linker database for entities.
func (e Static) Clean() error {
	return e.repo.DeleteAll()
}

// Get links for target.
func (e Static) Links(target shared.RemoteName) ([]Linked, error) {
	rem, ok := e.remotes[target]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(_packageName, target)
	}
	return rem.Linkables().Links()
}

// Get count of not matched entities.
func (e Static) NotMatchedCount(target shared.RemoteName) (int, error) {
	rem, ok := e.remotes[target]
	if !ok {
		return 0, shared.NewErrRemoteNotFound(_packageName, target)
	}
	return rem.Linkables().NotMatchedCount()
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

	theLog := _log.
		AddField("from", target.RemoteName().String()).
		AddField("name", target.Name()).
		AddField("remoteID", target.ID().String())

	theLog.Info("FIND ENTITY FOR")

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
		return nil, shared.NewErrRemoteNotFound(_packageName, target)
	}

	theLog := _log.
		AddField("from", who.RemoteName().String()).
		AddField("name", who.Name()).
		AddField("remoteID", who.ID().String()).
		AddField("to", target.String())

	theLog.Info("==== 🔎 ====")

	// same remotes.
	if target == who.RemoteName() {
		theLog.Info("✅ (same remotes)")
		return who, nil
	}

	// Match.
	matched, err := targetRem.Match(theLog.WithContext(ctx), who)
	if err != nil {
		return nil, err
	}
	if shared.IsNil(matched) {
		theLog.Info("❌")
		return nil, err
	}

	theLog.AddField("matchedRemoteID", matched.ID().String()).Info("✅")

	return matched, err
}
