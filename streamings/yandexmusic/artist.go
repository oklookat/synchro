package yandexmusic

import (
	"context"
	"sort"

	"github.com/oklookat/goym"
	"github.com/oklookat/goym/schema"
)

func newArtist(artist schema.Artist, client *goym.Client) (*Artist, error) {
	if len(artist.ID) == 0 {
		return nil, errEmptyID
	}
	return &Artist{
		Entity: newEntity(artist.ID.String(), artist.Name),
		client: client,
		artist: artist,
	}, nil
}

type Artist struct {
	*Entity
	artist schema.Artist
	client *goym.Client

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
	totalCurrent := 0
	page := 0

	for {
		result, err := e.client.ArtistAlbums(ctx, e.artist.ID, page, 35, schema.SortByYear, schema.SortOrderAsc)
		if err != nil {
			if isNotFound(err) {
				e.isCachedOldestAlbumsNames = true
				return e.cachedOldestAlbumsNames, nil
			}
			return e.cachedOldestAlbumsNames, err
		}
		if len(result.Result.Albums) == 0 {
			break
		}
		if result.Pager == nil {
			break
		}
		if totalCurrent >= result.Pager.Total {
			break
		}
		for _, alb := range result.Result.Albums {
			if alb.MetaType != schema.AlbumMetaTypeMusic {
				continue
			}
			allAlbums = append(allAlbums, alb)
		}
		totalCurrent += len(result.Result.Albums)
		page++
	}

	if len(allAlbums) == 0 {
		e.isCachedOldestAlbumsNames = true
		return e.cachedOldestAlbumsNames, nil
	}

	// Oldest first.
	sort.SliceStable(allAlbums, func(i, j int) bool {
		return allAlbums[i].ReleaseDate.Before(allAlbums[j].ReleaseDate)
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

	var allSingles []schema.Album
	page := 0
	totalCurrent := 0

	for {
		albums, err := e.client.ArtistAlbums(ctx, e.artist.ID, 0, 35, schema.SortByYear, schema.SortOrderAsc)
		if err != nil {
			if isNotFound(err) {
				e.isCachedOldestSinglesNames = true
				return e.cachedOldestSinglesNames, nil
			}
			return e.cachedOldestSinglesNames, err
		}
		if albums.Pager == nil {
			break
		}
		if totalCurrent >= albums.Pager.Total {
			break
		}
		for _, alb := range albums.Result.Albums {
			if alb.MetaType != schema.AlbumMetaTypeSingle {
				continue
			}
			allSingles = append(allSingles, alb)
		}
		totalCurrent += len(albums.Result.Albums)
		page++
	}

	if len(allSingles) == 0 {
		e.isCachedOldestSinglesNames = true
		return e.cachedOldestSinglesNames, nil
	}

	// Oldest first.
	sort.SliceStable(allSingles, func(i, j int) bool {
		return allSingles[i].ReleaseDate.Before(allSingles[j].ReleaseDate)
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
