package spotify

import (
	"context"
	"sort"

	"github.com/zmb3/spotify/v2"
)

func newArtist(artist spotify.SimpleArtist, client *spotify.Client) *Artist {
	return &Artist{
		Entity: newEntity(artist.ID.String(), artist.Name),
		client: client,
		artist: artist,
	}
}

type Artist struct {
	*Entity
	artist spotify.SimpleArtist
	client *spotify.Client

	isCachedOldestAlbumsNames bool
	cachedOldestAlbumsNames   [20]string

	isCachedOldestSinglesNames bool
	cachedOldestSinglesNames   [20]string
}

func (e *Artist) OldestAlbumsNames(ctx context.Context) ([20]string, error) {
	if e.isCachedOldestAlbumsNames {
		return e.cachedOldestAlbumsNames, nil
	}

	var allAlbums []spotify.SimpleAlbum

	// Get all albums.
	offset := 0
	for {
		page, err := e.client.GetArtistAlbums(
			ctx,
			e.artist.ID,
			[]spotify.AlbumType{spotify.AlbumTypeAlbum},
			_market,
			spotify.Limit(15),
			spotify.Offset(offset),
		)
		if err != nil {
			return e.cachedOldestAlbumsNames, err
		}
		if len(page.Albums) == 0 {
			break
		}
		allAlbums = append(allAlbums, page.Albums...)
		if len(page.Next) == 0 {
			break
		}
		offset += int(page.Limit)
	}

	// Oldest first.
	sort.SliceStable(allAlbums, func(i, j int) bool {
		return allAlbums[i].ReleaseDateTime().Before(allAlbums[j].ReleaseDateTime())
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

	var allSingles []spotify.SimpleAlbum

	// Get all single albums.
	offset := 0
	for {
		result, err := e.client.GetArtistAlbums(
			ctx, e.artist.ID,
			[]spotify.AlbumType{spotify.AlbumTypeSingle},
			_market,
			spotify.Limit(15),
			spotify.Offset(offset),
		)
		if err != nil {
			return e.cachedOldestSinglesNames, err
		}
		if result == nil {
			break
		}
		allSingles = append(allSingles, result.Albums...)
		if len(result.Next) == 0 {
			break
		}
		offset += int(result.Limit)
	}

	// Oldest first.
	sort.SliceStable(allSingles, func(i, j int) bool {
		return allSingles[i].ReleaseDateTime().Before(allSingles[j].ReleaseDateTime())
	})

	for i := range e.cachedOldestSinglesNames {
		if i == len(allSingles) {
			break
		}
		e.cachedOldestSinglesNames[i] = allSingles[i].Name
	}

	e.isCachedOldestSinglesNames = true
	return e.cachedOldestSinglesNames, nil
}
