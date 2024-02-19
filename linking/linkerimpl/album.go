package linkerimpl

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
)

func NewAlbums() (*linker.Static, error) {
	ready := checkRemotes()
	if len(ready) == 0 {
		return nil, shared.NewErrNoAvailableRemotes(_packageName)
	}

	converted := map[shared.RemoteName]linker.Remote{}
	for name := range ready {
		converted[name] = AlbumsRemote{repo: ready[name].Repository()}
	}

	return linker.NewStatic(repository.AlbumEntity, converted), nil
}

type AlbumsRemote struct {
	repo shared.RemoteRepository
}

func (e AlbumsRemote) Name() shared.RemoteName {
	return e.repo.Name()
}

func (e AlbumsRemote) RemoteEntity(ctx context.Context, id shared.RemoteID) (linker.RemoteEntity, error) {
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	al, err := actions.Album(ctx, id)
	return al, err
}

func (e AlbumsRemote) Linkables() linker.Linkables {
	return repository.NewLinkableEntity(repository.EntityNameAlbum, e.repo.Name())
}

func (e AlbumsRemote) Match(ctx context.Context, target linker.RemoteEntity) (linker.RemoteEntity, error) {
	realTarget, ok := target.(shared.RemoteAlbum)
	if !ok {
		return nil, errors.New("realTarget, ok := target.(shared.RemoteAlbum)")
	}

	// if same remotes
	if e.repo.Name() == realTarget.RemoteName() {
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
