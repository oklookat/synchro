package deezer

import (
	"context"
	"sort"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/schema"
)

func newArtist(cl *deezus.Client, ent schema.SimpleArtist) *Artist {
	return &Artist{
		Entity: newEntity(ent.ID.String(), ent.Name),
		client: cl,
		artist: ent,
	}
}

type Artist struct {
	*Entity
	artist schema.SimpleArtist
	client *deezus.Client

	isCachedOldestAlbumsNames bool
	cachedOldestAlbumsNames   [20]string

	isCachedOldestSinglesNames bool
	cachedOldestSinglesNames   [20]string
}

func (e *Artist) OldestAlbumsNames(ctx context.Context) ([20]string, error) {
	if e.isCachedOldestAlbumsNames {
		return e.cachedOldestAlbumsNames, nil
	}

	var allAlbums []schema.SimpleAlbum

	// Get all albums.
	offset := 0
	const limit = 30
	for {
		resp, err := e.client.ArtistAlbums(
			ctx,
			e.artist.ID,
			offset,
			limit,
		)
		if err != nil {
			return e.cachedOldestAlbumsNames, err
		}
		if len(resp.Data) == 0 {
			break
		}
		allAlbums = append(allAlbums, resp.Data...)
		if resp.Next == nil || len(*resp.Next) == 0 {
			break
		}
		offset += limit
	}

	// Oldest first.
	sort.SliceStable(allAlbums, func(i, j int) bool {
		if allAlbums[i].ReleaseDate == nil || allAlbums[j].ReleaseDate == nil {
			return false
		}
		return allAlbums[i].ReleaseDate.Time().Before(allAlbums[j].ReleaseDate.Time())
	})

	for i := range e.cachedOldestAlbumsNames {
		if i == len(allAlbums) {
			break
		}
		e.cachedOldestAlbumsNames[i] = allAlbums[i].Title
	}

	e.isCachedOldestAlbumsNames = true
	return e.cachedOldestAlbumsNames, nil
}

func (e *Artist) OldestSinglesNames(ctx context.Context) ([20]string, error) {
	if e.isCachedOldestSinglesNames {
		return e.cachedOldestSinglesNames, nil
	}

	var allSingles []schema.SimpleAlbum

	// Get all single albums.
	offset := 0
	const limit = 30
	for {
		result, err := e.client.ArtistAlbums(
			ctx, e.artist.ID,
			offset,
			limit,
		)
		if err != nil {
			return e.cachedOldestSinglesNames, err
		}
		if result == nil {
			break
		}
		allSingles = append(allSingles, result.Data...)
		if result.Next == nil || len(*result.Next) == 0 {
			break
		}
		offset += limit
	}

	// Oldest first.
	sort.SliceStable(allSingles, func(i, j int) bool {
		if allSingles[i].ReleaseDate == nil || allSingles[j].ReleaseDate == nil {
			return false
		}
		return allSingles[i].ReleaseDate.Time().Before(allSingles[j].ReleaseDate.Time())
	})

	for i := range e.cachedOldestSinglesNames {
		if i == len(allSingles) {
			break
		}
		e.cachedOldestSinglesNames[i] = allSingles[i].Title
	}

	e.isCachedOldestSinglesNames = true
	return e.cachedOldestSinglesNames, nil
}
