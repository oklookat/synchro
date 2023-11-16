package linkerimpl

import (
	"context"
	"errors"
	"strings"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

func NewArtists() (*linker.Static, error) {
	ready := checkRemotes()
	if len(ready) == 0 {
		return nil, errors.New("no services")
	}

	converted := map[streaming.ServiceName]linker.Service{}
	for name := range ready {
		converted[name] = ArtistsRemote{repo: ready[name].Database()}
	}

	return linker.NewStatic(repository.EntityArtist, converted), nil
}

type ArtistsRemote struct {
	repo streaming.Database
}

func (e ArtistsRemote) Name() streaming.ServiceName {
	return e.repo.Name()
}

func (e ArtistsRemote) StreamingServiceEntity(ctx context.Context, id streaming.ServiceEntityID) (linker.StreamingServiceEntity, error) {
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	ar, err := actions.Artist(ctx, id)
	return ar, err
}

func (e ArtistsRemote) Linkables() linker.Linkables {
	real, ok := e.repo.(*repository.Service)
	if !ok {
		return nil
	}
	return repository.NewLinkableEntityArtist(real)
}

func (e ArtistsRemote) Match(ctx context.Context, target linker.StreamingServiceEntity) (linker.StreamingServiceEntity, error) {
	realTarget, ok := target.(streaming.ServiceArtist)
	if !ok {
		return nil, errors.New("realTarget, ok := target.(streaming.ServiceArtist)")
	}

	// If same services.
	if e.repo.Name() == realTarget.ServiceName() {
		return target, nil
	}

	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}

	oldestAlbumsNames, err := realTarget.OldestAlbumsNames(ctx)
	if err != nil {
		return nil, err
	}
	oldestSinglesNames, err := realTarget.OldestSinglesNames(ctx)
	if err != nil {
		return nil, err
	}
	normalizedOldestAlbumsNames := shared.NormalizeStringSliceSearchablePart(oldestAlbumsNames[:])
	normalizedOldestSinglesNames := shared.NormalizeStringSliceSearchablePart(oldestSinglesNames[:])

	var candidate artistCandidate
	searchResult, err := actions.SearchArtists(ctx, realTarget)
	if err != nil {
		return nil, err
	}
	for i := range searchResult {
		if shared.IsNil(searchResult[i]) {
			break
		}

		// If target dont have albums and singles.
		if len(oldestAlbumsNames) == 0 && len(oldestSinglesNames) == 0 {
			return searchResult[i], err
		}

		// Albums.
		fAlbumsNames, err := searchResult[i].OldestAlbumsNames(ctx)
		if err != nil {
			return nil, err
		}
		nAlbumsNames := shared.NormalizeStringSliceSearchablePart(fAlbumsNames[:])
		albumsWeight := shared.SameNameSlices(normalizedOldestAlbumsNames, nAlbumsNames)

		// Singles.
		fSinglesNames, err := searchResult[i].OldestSinglesNames(ctx)
		if err != nil {
			return nil, err
		}
		nSinglesNames := shared.NormalizeStringSliceSearchablePart(fSinglesNames[:])
		singlesWeight := shared.SameNameSlices(normalizedOldestSinglesNames, nSinglesNames)

		// Total.
		total := albumsWeight + singlesWeight
		if total > candidate.weight {
			candidate.weight = total
			candidate.candidate = searchResult[i]
		}
	}

	if candidate.weight == 0 {
		if len(searchResult) > 0 {
			// Just compare first result by name.
			if strings.EqualFold(shared.Normalize(target.Name()), shared.Normalize(searchResult[0].Name())) {
				// POTENTIAL MISMATCH (compared by names only).
				return searchResult[0], err
			}
		}
		return nil, err
	}

	return candidate.candidate, err
}

type artistCandidate struct {
	weight    float64
	candidate streaming.ServiceArtist
}
