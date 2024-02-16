package vkmusic

import (
	"context"
	"sort"

	"github.com/oklookat/govkm"
	"github.com/oklookat/govkm/schema"
)

func newArtist(artist schema.SimpleArtist, client *govkm.Client) *Artist {
	return &Artist{
		Entity: newEntity(artist.APIID.String(), artist.Name),
		client: client,
		artist: artist,
	}
}

type Artist struct {
	*Entity
	artist schema.SimpleArtist
	client *govkm.Client

	isCachedOldestAlbumsNames bool
	cachedOldestAlbumsNames   [20]string

	isCachedOldestSinglesNames bool
	cachedOldestSinglesNames   [20]string
}

func (e *Artist) OldestAlbumsNames(ctx context.Context) ([20]string, error) {
	if e.isCachedOldestAlbumsNames {
		return e.cachedOldestAlbumsNames, nil
	}

	var allAlbums []schema.Album

	const limit = 30
	offset := 0
	for {
		resp, err := e.client.ArtistAlbums(ctx,
			e.artist.APIID,
			[]schema.AlbumType{schema.AlbumTypeAlbum},
			limit, offset)
		if err != nil {
			return e.cachedOldestAlbumsNames, err
		}
		if len(resp.Data.Albums) == 0 {
			break
		}
		allAlbums = append(allAlbums, resp.Data.Albums...)
		offset += limit
	}

	// Oldest first.
	sort.SliceStable(allAlbums, func(i, j int) bool {
		return allAlbums[i].ReleaseDateTimestamp.
			Time().
			Before(allAlbums[j].ReleaseDateTimestamp.Time())
	})

	for i := range e.cachedOldestAlbumsNames {
		if i == len(allAlbums) {
			break
		}
		e.cachedOldestAlbumsNames[i] = allAlbums[i].Name
	}

	e.isCachedOldestAlbumsNames = true
	return e.cachedOldestAlbumsNames, nil
}

func (e *Artist) OldestSinglesNames(ctx context.Context) ([20]string, error) {
	if e.isCachedOldestSinglesNames {
		return e.cachedOldestSinglesNames, nil
	}

	var allSingleAlbums []schema.Album

	const limit = 50
	offset := 0
	for {
		resp, err := e.client.ArtistAlbums(ctx,
			e.artist.APIID,
			[]schema.AlbumType{schema.AlbumTypeSingle},
			limit, offset)
		if err != nil {
			return e.cachedOldestSinglesNames, err
		}
		if len(resp.Data.Albums) == 0 {
			break
		}
		allSingleAlbums = append(allSingleAlbums, resp.Data.Albums...)
		offset += limit
	}

	// Oldest first.
	sort.SliceStable(allSingleAlbums, func(i, j int) bool {
		return allSingleAlbums[i].ReleaseDateTimestamp.
			Time().
			Before(allSingleAlbums[j].ReleaseDateTimestamp.Time())
	})

	for i := range e.cachedOldestSinglesNames {
		if i == len(allSingleAlbums) {
			break
		}
		e.cachedOldestSinglesNames[i] = allSingleAlbums[i].Name
	}

	e.isCachedOldestSinglesNames = true
	return e.cachedOldestSinglesNames, nil
}
