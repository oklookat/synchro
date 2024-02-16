package linkerimpl

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
)

func NewTracks() (*linker.Static, error) {
	ready := checkRemotes()
	if len(ready) == 0 {
		return nil, shared.NewErrNoAvailableRemotes(_packageName)
	}

	converted := map[shared.RemoteName]linker.Remote{}
	for name := range ready {
		converted[name] = TracksRemote{repo: ready[name].Repository()}
	}

	return linker.NewStatic(repository.EntityTrack{}, converted), nil
}

type TracksRemote struct {
	repo shared.RemoteRepository
}

func (e TracksRemote) Name() shared.RemoteName {
	return e.repo.Name()
}

func (e TracksRemote) RemoteEntity(ctx context.Context, id shared.RemoteID) (linker.RemoteEntity, error) {
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	return actions.Track(ctx, id)
}

func (e TracksRemote) Linkables() linker.Linkables {
	real, ok := e.repo.(*repository.Remote)
	if !ok {
		_log.Error("real, ok := e.repo.(*repository.Remote)")
		return nil
	}
	return repository.NewLinkableTrack(real)
}

func (e TracksRemote) Match(ctx context.Context, target linker.RemoteEntity) (linker.RemoteEntity, error) {
	realTarget, ok := target.(shared.RemoteTrack)
	if !ok {
		return nil, errors.New("realTarget, ok := target.(shared.RemoteTrack)")
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
	tracks, err := actions.SearchTracks(ctx, realTarget)
	if err != nil {
		return nil, err
	}

	// Match.
	matched := matchTrack(realTarget, tracks[:])
	if shared.IsNil(matched) {
		return nil, nil
	}

	return matched, err
}
