package linkerimpl

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
)

func NewArtists() (*linker.Static, error) {
	ready := checkRemotes()
	if len(ready) == 0 {
		return nil, shared.NewErrNoAvailableRemotes()
	}

	converted := map[shared.RemoteName]linker.Remote{}
	for name := range ready {
		converted[name] = ArtistsRemote{repo: ready[name].Repository()}
	}

	return linker.NewStatic(repository.ArtistEntity, converted), nil
}

type ArtistsRemote struct {
	repo shared.RemoteRepository
}

func (e ArtistsRemote) Name() shared.RemoteName {
	return e.repo.Name()
}

func (e ArtistsRemote) RemoteEntity(ctx context.Context, id shared.RemoteID) (linker.RemoteEntity, error) {
	actions, err := e.repo.Actions()
	if err != nil {
		return nil, err
	}
	ar, err := actions.Artist(ctx, id)
	return ar, err
}

func (e ArtistsRemote) Linkables() linker.Linkables {
	return repository.NewLinkableEntity(repository.EntityNameArtist, e.repo.Name())
}

func (e ArtistsRemote) Match(ctx context.Context, target linker.RemoteEntity) (linker.RemoteEntity, error) {
	realTarget, ok := target.(shared.RemoteArtist)
	if !ok {
		return nil, errors.New("realTarget, ok := target.(shared.RemoteArtist)")
	}

	// If same remotes.
	if e.repo.Name() == realTarget.RemoteName() {
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
				slog.Warn("POTENTIAL MISMATCH (compared by names only)")
				return searchResult[0], err
			}
		}
		return nil, err
	}

	return candidate.candidate, err
}

type artistCandidate struct {
	weight    float64
	candidate shared.RemoteArtist
}
