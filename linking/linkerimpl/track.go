package linkerimpl

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

func NewTracks() (*linker.Static, error) {
	ready := checkRemotes()
	if len(ready) == 0 {
		return nil, errors.New("no services")
	}

	converted := map[streaming.ServiceName]linker.Service{}
	for name := range ready {
		converted[name] = TracksRemote{repo: ready[name].Database()}
	}

	return linker.NewStatic(repository.EntityTrack, converted), nil
}

type TracksRemote struct {
	repo streaming.Database
}

func (e TracksRemote) Name() streaming.ServiceName {
	return e.repo.Name()
}

func (e TracksRemote) StreamingServiceEntity(ctx context.Context, id streaming.ServiceEntityID) (linker.StreamingServiceEntity, error) {
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	return actions.Track(ctx, id)
}

func (e TracksRemote) Linkables() linker.Linkables {
	real, ok := e.repo.(*repository.Service)
	if !ok {
		return nil
	}
	return repository.NewLinkableEntityTrack(real)
}

func (e TracksRemote) Match(ctx context.Context, target linker.StreamingServiceEntity) (linker.StreamingServiceEntity, error) {
	realTarget, ok := target.(streaming.ServiceTrack)
	if !ok {
		return nil, errors.New("realTarget, ok := target.(streaming.ServiceTrack)")
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
