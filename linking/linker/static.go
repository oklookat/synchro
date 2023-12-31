package linker

import (
	"context"
	"errors"
	"time"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

type (
	// Repository of service-independent entity.
	//
	// Example: artist repository.
	Repository interface {
		CreateEntity() (streaming.DatabaseEntityID, error)

		// 1. Delete entities that not linked with any service.
		//
		// 2. Delete entities that linked, but have NULL RemoteID on all services.
		DeleteNotLinked() error
	}

	// Links global entities like artists, tracks, albums.
	Service interface {
		// Unique name.
		//
		// Example: "Spotify".
		Name() streaming.ServiceName

		// Find entity from another service in current.
		Match(context.Context, StreamingServiceEntity) (StreamingServiceEntity, error)

		// Get entity by ID.
		StreamingServiceEntity(context.Context, streaming.ServiceEntityID) (StreamingServiceEntity, error)

		// DB ops for linked entities.
		Linkables() Linkables
	}

	// Service in DB.
	Linkables interface {
		// Link entity with service.
		CreateLink(context.Context, streaming.DatabaseEntityID, *streaming.ServiceEntityID) (Linked, error)

		// Example: get linked spotify artist by artist entity.
		LinkedEntity(streaming.DatabaseEntityID) (Linked, error)

		// Example: get linked spotify artist by its ID on Spotify.
		LinkedRemoteID(streaming.ServiceEntityID) (Linked, error)
	}

	StreamingServiceEntity interface {
		// Example: Spotify.
		ServiceName() streaming.ServiceName

		// Example: Spotify artist ID.
		ID() streaming.ServiceEntityID

		// Example: Spotify artist name.
		Name() string
	}

	Linked interface {
		// Parent entity.
		EntityID() streaming.DatabaseEntityID

		// Example: Spotify artist ID.
		//
		// Nil if not exists in service.
		ServiceEntityID() *streaming.ServiceEntityID

		// Example: set Spotify artist ID.
		//
		// Nil if not exists in service.
		SetServiceEntityID(*streaming.ServiceEntityID) error

		// Date when link created/modified.
		ModifiedAt() time.Time
	}

	ToStreamingResult struct {
		// Missing on service before but now present?
		MissingBefore,

		// Missing now.
		MissingNow,

		// Link not exists before?
		NewLink bool

		// Entity linked with target service.
		//
		// Can be nil.
		Linked Linked
	}

	FromStreamingResult struct {
		// Service entity from target.
		StreamingServiceEntity StreamingServiceEntity

		// Linked.
		Linked Linked
	}
)

func NewStatic(repo Repository, services map[streaming.ServiceName]Service) *Static {
	return &Static{
		repo:     repo,
		services: services,
	}
}

// Links global entities.
//
// Example: track, artist, album.
type Static struct {
	repo     Repository
	services map[streaming.ServiceName]Service
}

// From service entity to linked.
func (e Static) FromStreaming(ctx context.Context, target StreamingServiceEntity) (FromStreamingResult, error) {
	log := _log.With().Str("func", "FromStreaming").Stringer("targetService", target.ServiceName()).Str("targetName", target.Name()).Stringer("targetID", target.ID()).Logger()

	result := FromStreamingResult{}
	result.StreamingServiceEntity = target

	targetRemote, ok := e.services[target.ServiceName()]
	if !ok {
		return result, shared.NewErrServiceNotFound(target.ServiceName())
	}

	// Link exists?
	linked, err := targetRemote.Linkables().LinkedRemoteID(target.ID())
	if err != nil {
		return result, err
	}

	// Link exists.
	if !shared.IsNil(linked) {
		// Missing before?
		if linked.ServiceEntityID() == nil {
			// Set ID.
			log.Info().Msg("Missing before, set ID")
			updId := target.ID()
			if err = linked.SetServiceEntityID(&updId); err != nil {
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

	// Not linked by all services.
	if entitiesResult.FoundEntity == nil {
		log.Info().Msg("Not linked, create entity")
		// Create new entity.
		entityId, err := e.repo.CreateEntity()
		if err != nil {
			return result, err
		}
		entitiesResult.FoundEntity = &entityId
	}
	entityId := *entitiesResult.FoundEntity

	// Link entity with services.
	for remName, remId := range entitiesResult.TargetRemoteId {
		// Get service.
		service := e.services[remName]

		// Link exists?
		linked, err := service.Linkables().LinkedEntity(entityId)
		if err != nil {
			return result, err
		}

		// Not exists.
		if shared.IsNil(linked) {
			log.Info().Msg("Create link")
			// Link.
			newLinked, err := service.Linkables().CreateLink(ctx, entityId, remId)
			if err != nil {
				return result, err
			}
			linked = newLinked
			// Add to result if current service is a target.
			if remName == target.ServiceName() {
				result.Linked = linked
			}
		}

		// Exists.

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

		idNotChanged := ((linked.ServiceEntityID() != nil && remId != nil) &&
			(*linked.ServiceEntityID() == *remId))
		if !idNotChanged {
			log.Info().Stringer("oldID", linked.ServiceEntityID()).Msg("Overwrite ID")
			if err := linked.SetServiceEntityID(remId); err != nil {
				return result, err
			}
		}
	}

	return result, err
}

// From entity to service entity.
func (e Static) ToStreaming(ctx context.Context, id streaming.DatabaseEntityID, target streaming.ServiceName) (ToStreamingResult, error) {
	log := _log.With().Str("func", "ToStreaming").Stringer("databaseEntityID", id).Stringer("target", target).Logger()

	result := ToStreamingResult{}

	targetRem, ok := e.services[target]
	if !ok {
		return result, shared.NewErrServiceNotFound(target)
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
		result.MissingBefore = linked.ServiceEntityID() == nil
		result.MissingNow = result.MissingBefore
		if !result.MissingBefore {
			// Not missing.
			return result, err
		}

		// Recheck? Maybe it's not missing now.

		cfg := &Config{}
		if err := config.Get(cfg.Key(), cfg); err != nil {
			return result, err
		}

		if !cfg.RecheckMissing {
			// Recheck disabled in config.
			log.Info().Msg("Not found, recheck disabled in config")
			return result, err
		}
	}

	log.Info().Msg("Recheck")

	// Entity not linked with target or missing before.
	// Create/find link.

	// Find link in another services.
	for name := range e.services {
		// Skip current.
		if name == target {
			continue
		}

		log2 := log.With().Stringer("inService", name).Logger()
		log2.Info().Msg("Create/find link in")

		// Link exists?
		linkedFromAnother, err := e.services[name].Linkables().LinkedEntity(id)
		if err != nil {
			return result, err
		}

		// Not exists?
		if shared.IsNil(linkedFromAnother) {
			// Try another.
			continue
		}

		// Exists, but missing.
		if linkedFromAnother.ServiceEntityID() == nil {
			// Skip.
			continue
		}

		// Exists, and not missing. Try to get entity in service.
		entityFromAnotherRemote, err := e.services[name].StreamingServiceEntity(ctx, *linkedFromAnother.ServiceEntityID())
		if err != nil {
			return result, err
		}

		// Not exists?
		if shared.IsNil(entityFromAnotherRemote) {
			// Make missing.
			if err = linkedFromAnother.SetServiceEntityID(nil); err != nil {
				return result, err
			}
			// Try another.
			continue
		}

		// Exists. Search entity from another service in target.
		found, err := e.search(ctx, entityFromAnotherRemote, target)
		if err != nil {
			return result, err
		}

		// Not found?
		if shared.IsNil(found) {
			log2.Info().Msg("Stay missing")

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

		log2.Info().Stringer("foundID", foundID).Msg("Found, missing before")

		// Link exists?
		if !shared.IsNil(result.Linked) {
			// Set ID.
			return result, linked.SetServiceEntityID(&foundID)
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
	Target StreamingServiceEntity

	// Entity that can be linked with Target.
	FoundEntity *streaming.DatabaseEntityID

	// 1. Target ID in another services that also not linked with FoundEntities.
	// Includes Target.
	//
	// 2. Target ID in another services that missing in service.
	TargetRemoteId map[streaming.ServiceName]*streaming.ServiceEntityID
}

// Find an entities to link with target.
func (e Static) findEntitiesToLink(ctx context.Context, target StreamingServiceEntity) (findEntitiesToLinkResult, error) {
	log := _log.With().Str("func", "findEntitiesToLink").Stringer("fromService", target.ServiceName()).Str("targetName", target.Name()).Stringer("targetID", target.ID()).Logger()

	result := findEntitiesToLinkResult{
		Target:         target,
		TargetRemoteId: map[streaming.ServiceName]*streaming.ServiceEntityID{},
	}

	// Add current.
	targetId := target.ID()
	result.TargetRemoteId[target.ServiceName()] = &targetId

	for _, service := range e.services {
		// Skip current.
		if service.Name() == target.ServiceName() {
			continue
		}

		log2 := log.With().Stringer("inService", service.Name()).Logger()

		// Find target.
		log2.Info().Msg("==== 🔎 ====")
		foundTarget, err := e.search(ctx, target, service.Name())
		if err != nil {
			return result, err
		}

		// Missing?
		if shared.IsNil(foundTarget) {
			log2.Info().Msg("==== ❌ ====")
			// Mark as missing.
			result.TargetRemoteId[service.Name()] = nil
			continue
		}

		// Found.
		log2.Info().Str("foundName", foundTarget.Name()).Stringer("foundID", foundTarget.ID()).Msg("==== ✅ ====")

		// Linked?
		linked, err := service.Linkables().LinkedRemoteID(foundTarget.ID())
		if err != nil {
			return result, err
		}

		// Not linked.
		if shared.IsNil(linked) {
			foundID := foundTarget.ID()
			result.TargetRemoteId[service.Name()] = &foundID
			continue
		}

		// Linked. Get entity id.
		entityId := linked.EntityID()
		result.FoundEntity = &entityId
		return result, err
	}

	return result, nil
}

// Search any service entity in any service.
//
// Example: search Spotify artist in Yandex.Music.
//
// Returns service entity from target.
//
// Nil if not exists.
func (e Static) search(ctx context.Context, who StreamingServiceEntity, target streaming.ServiceName) (StreamingServiceEntity, error) {
	targetRem, ok := e.services[target]
	if !ok {
		return nil, shared.NewErrServiceNotFound(target)
	}

	// same services.
	if target == who.ServiceName() {
		return who, nil
	}

	// Match.
	matched, err := targetRem.Match(ctx, who)
	if err != nil {
		return nil, err
	}
	if shared.IsNil(matched) {
		return nil, err
	}
	return matched, err
}
