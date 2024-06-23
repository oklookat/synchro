package linker

import (
	"context"
	"log/slog"

	"github.com/oklookat/synchro/shared"
)

type (
	RepositoryDynamic interface {
		Repository

		// Delete entity by id.
		DeleteEntity(shared.EntityID) error
	}

	RemoteDynamic interface {
		// Unique name.
		//
		// Example: unique account ID in DB.
		Name() shared.RemoteName

		// Example: create playlist with name.
		Create(context.Context, string) (RemoteEntity, error)

		// DB ops for linked entities.
		Linkables() LinkablesDynamic
	}

	// Remote in DB.
	LinkablesDynamic interface {
		// Link entity with ren.
		CreateLink(context.Context, shared.EntityID, shared.RemoteID) (LinkedDynamic, error)

		// Example: get linked spotify artist by artist entity.
		LinkedEntity(shared.EntityID) (LinkedDynamic, error)

		// Example: get linked spotify artist by its ID on Spotify.
		LinkedRemoteID(shared.RemoteID) (LinkedDynamic, error)
	}

	LinkedDynamic interface {
		// Parent entity.
		EntityID() shared.EntityID

		// Example: Spotify playlist ID.
		RemoteID() shared.RemoteID
	}
)

func NewDynamic(repo RepositoryDynamic, remotes map[shared.RemoteName]RemoteDynamic) *Dynamic {
	slog.Info("lovesYou", "linker.Dynamic", "~~~ LET'S CREATE SOME THINGS! ~~~")
	return &Dynamic{
		repo:    repo,
		remotes: remotes,
	}
}

type Dynamic struct {
	repo    RepositoryDynamic
	remotes map[shared.RemoteName]RemoteDynamic
}

// From remote entity to linked.
func (e Dynamic) FromRemote(ctx context.Context, target RemoteEntity) (LinkedDynamic, RemoteEntity, error) {
	rem, ok := e.remotes[target.RemoteName()]
	if !ok {
		return nil, nil, shared.NewErrRemoteNotFound(target.RemoteName())
	}

	// Link exists?
	linked, err := rem.Linkables().LinkedRemoteID(target.ID())
	if err != nil {
		return nil, nil, err
	}

	// Exists.
	if !shared.IsNil(linked) {
		return linked, target, err
	}

	// Not exists, create links.
	linked, err = e.createLinks(ctx, target.RemoteName(), target.ID(), target.Name())
	return linked, target, err
}

// From local entity to Linked.
func (e Dynamic) ToRemote(ctx context.Context, id shared.EntityID, target shared.RemoteName, createIfNot bool) (LinkedDynamic, error) {
	targetRem, ok := e.remotes[target]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(target)
	}

	// Link exists?
	linked, err := targetRem.Linkables().LinkedEntity(id)
	if err != nil {
		return nil, err
	}

	// Exists.
	if !shared.IsNil(linked) || !createIfNot {
		return linked, err
	}

	// Not exists, create links.
	name := "SYNCHRO_" + shared.GenerateWord()
	created, err := targetRem.Create(ctx, name)
	if err != nil {
		return linked, err
	}

	linked, err = e.createLinks(ctx, target, created.ID(), name)
	return linked, err
}

// Delete linked entity with entity & delete remote entities in remotes.
func (e Dynamic) RemoveLinks(ctx context.Context, ids map[shared.EntityID]bool) error {
	if len(ids) == 0 {
		return nil
	}
	for eId := range ids {
		for name := range e.remotes {
			linkd, err := e.remotes[name].Linkables().LinkedEntity(eId)
			if err != nil {
				return err
			}
			if shared.IsNil(linkd) {
				continue
			}
			if err := e.repo.DeleteEntity(linkd.EntityID()); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// Clean linker database for this entity.
func (e Dynamic) Clean() error {
	return e.repo.DeleteAll()
}

func (e Dynamic) createLinks(ctx context.Context, target shared.RemoteName, origin shared.RemoteID, originName string) (LinkedDynamic, error) {
	// Get current rem/account.
	rem, ok := e.remotes[target]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(target)
	}
	defer e.repo.DeleteNotLinked()

	// Create entity.
	eId, err := e.repo.CreateEntity()
	if err != nil {
		return nil, err
	}

	// Link with current.
	linked, err := rem.Linkables().CreateLink(ctx, eId, origin)
	if err != nil {
		return linked, err
	}

	// Create remote entities in another remotes, link with created entity.
	for name := range e.remotes {
		if name == target {
			continue
		}
		created, err := e.remotes[name].Create(ctx, originName)
		if err != nil {
			return linked, err
		}
		_, err = e.remotes[name].Linkables().CreateLink(ctx, eId, created.ID())
		if err != nil {
			return linked, err
		}
	}

	return linked, err
}
