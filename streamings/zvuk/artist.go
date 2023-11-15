package zvuk

import (
	"context"
	"sort"

	"github.com/oklookat/gozvuk"
	"github.com/oklookat/gozvuk/schema"
)

func newArtist(artist *schema.SimpleArtist, client *gozvuk.Client) *Artist {
	if client == nil {
		return nil
	}
	ar := &Artist{
		Entity: newEntity(artist.ID.String(), artist.Title),
		client: client,
		artist: artist,
	}
	return ar
}

type Artist struct {
	*Entity
	artist *schema.SimpleArtist
	client *gozvuk.Client

	cachedArtist *schema.Artist

	isCachedOldestAlbumsNames bool
	cachedOldestAlbumsNames   [20]string

	isCachedOldestSinglesNames bool
	cachedOldestSinglesNames   [20]string
}

func (e *Artist) OldestAlbumsNames(ctx context.Context) ([20]string, error) {
	if err := e.cache(ctx); err != nil {
		return e.cachedOldestAlbumsNames, err
	}
	return e.cachedOldestAlbumsNames, nil
}

func (e *Artist) OldestSinglesNames(ctx context.Context) ([20]string, error) {
	if err := e.cache(ctx); err != nil {
		return e.cachedOldestSinglesNames, err
	}
	return e.cachedOldestSinglesNames, nil
}

func (e *Artist) cache(ctx context.Context) error {
	if e.cachedArtist != nil &&
		e.isCachedOldestAlbumsNames &&
		e.isCachedOldestSinglesNames {
		return nil
	}

	const limit = 20
	var (
		offset int

		allSingles []schema.SimpleRelease
		allAlbums  []schema.SimpleRelease
	)

	for {
		resp, err := e.client.GetArtists(ctx, []schema.ID{e.artist.ID}, true, limit, offset, false, 1, 0, false, 1, false)
		if err != nil {
			return err
		}
		if len(resp.Data.GetArtists) == 0 || len(resp.Data.GetArtists[0].ID) == 0 {
			break
		}
		if len(resp.Data.GetArtists[0].Releases) == 0 {
			break
		}
		if e.cachedArtist == nil {
			e.cachedArtist = &resp.Data.GetArtists[0]
		}
		for _, release := range resp.Data.GetArtists[0].Releases {
			if release.Type == schema.ReleaseTypeAlbum {
				allAlbums = append(allAlbums, release)
			}
			if release.Type == schema.ReleaseTypeSingle {
				allSingles = append(allSingles, release)
			}
		}
		offset += limit
	}

	// Oldest albums.
	sort.SliceStable(allAlbums, func(i, j int) bool {
		return allAlbums[i].Date.Before(allAlbums[j].Date.Time)
	})
	for i := range e.cachedOldestAlbumsNames {
		if i == len(allAlbums) {
			break
		}
		e.cachedOldestAlbumsNames[i] = allAlbums[i].Title
	}
	e.isCachedOldestAlbumsNames = true

	// Oldest singles.
	sort.SliceStable(allSingles, func(i, j int) bool {
		return allSingles[i].Date.Before(allSingles[j].Date.Time)
	})
	for i := range e.cachedOldestSinglesNames {
		if i == len(allSingles) {
			break
		}
		e.cachedOldestSinglesNames[i] = allSingles[i].Title
	}
	e.isCachedOldestSinglesNames = true

	return nil
}
