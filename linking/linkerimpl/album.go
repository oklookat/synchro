package linkerimpl

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

func NewAlbums() (*linker.Static, error) {
	ready := checkRemotes()
	if len(ready) == 0 {
		return nil, errors.New("no services")
	}

	converted := map[streaming.ServiceName]linker.Service{}
	for name := range ready {
		converted[name] = AlbumsRemote{repo: ready[name].Database()}
	}

	return linker.NewStatic(repository.EntityAlbum, converted), nil
}

type AlbumsRemote struct {
	repo streaming.Database
}

func (e AlbumsRemote) Name() streaming.ServiceName {
	return e.repo.Name()
}

func (e AlbumsRemote) StreamingServiceEntity(ctx context.Context, id streaming.ServiceEntityID) (linker.StreamingServiceEntity, error) {
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	al, err := actions.Album(ctx, id)
	return al, err
}

func (e AlbumsRemote) Linkables() linker.Linkables {
	real, ok := e.repo.(*repository.Service)
	if !ok {
		return nil
	}
	return repository.NewLinkableEntityAlbum(real)
}

func (e AlbumsRemote) Match(ctx context.Context, target linker.StreamingServiceEntity) (linker.StreamingServiceEntity, error) {
	realTarget, ok := target.(streaming.ServiceAlbum)
	if !ok {
		return nil, errors.New("realTarget, ok := target.(streaming.ServiceAlbum)")
	}

	// if same services
	if e.repo.Name() == realTarget.ServiceName() {
		return target, nil
	}

	// Search in target.
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	albums, err := actions.SearchAlbums(ctx, realTarget)
	if err != nil {
		return nil, err
	}

	// Match.
	matched := matchAlbum(realTarget, albums[:])
	if shared.IsNil(matched) {
		return nil, nil
	}

	return matched, nil
}
