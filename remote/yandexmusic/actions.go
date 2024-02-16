package yandexmusic

import (
	"context"

	"github.com/oklookat/goym"
	"github.com/oklookat/goym/schema"
	"github.com/oklookat/synchro/shared"
)

func newActions(client *goym.Client) *Actions {
	return &Actions{
		client: client,
	}
}

type Actions struct {
	client *goym.Client
}

func (e Actions) Album(ctx context.Context, id shared.RemoteID) (shared.RemoteAlbum, error) {
	resp, err := e.client.Album(ctx, schema.ID(id), false)
	notFound, err := isNotFoundOrErr(err, resp.Result.Title)
	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}
	return newAlbum(resp.Result, e.client)
}

func (e Actions) Artist(ctx context.Context, id shared.RemoteID) (shared.RemoteArtist, error) {
	resp, err := e.client.ArtistInfo(ctx, schema.ID(id))
	notFound, err := isNotFoundOrErr(err, resp.Result.Artist.Name)
	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}
	return newArtist(resp.Result.Artist.Artist, e.client)
}

func (e Actions) Track(ctx context.Context, id shared.RemoteID) (shared.RemoteTrack, error) {
	resp, err := e.client.Track(ctx, schema.ID(id))
	if err != nil {
		return nil, err
	}
	if len(resp.Result) == 0 {
		return nil, nil
	}
	return newTrack(resp.Result[0], e.client)
}

func (e Actions) SearchAlbums(ctx context.Context, what shared.RemoteAlbum) ([10]shared.RemoteAlbum, error) {
	action := &AlbumsSearchAction{e.client}
	return action.Search(ctx, what)
}

func (e Actions) SearchArtists(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	action := &ArtistsSearchAction{e.client}
	return action.Search(ctx, what)
}

func (e Actions) SearchTracks(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	action := &TracksSearchAction{e.client}
	return action.Search(ctx, what)
}

type AlbumsSearchAction struct {
	client *goym.Client
}

func (e AlbumsSearchAction) Search(
	ctx context.Context,
	what shared.RemoteAlbum,
) ([10]shared.RemoteAlbum, error) {

	artistName := what.Artists()[0].Name()
	albumName := what.Name()

	query := artistName + " " + shared.SearchablePart(albumName)
	result, err := e.search(ctx, query, false)
	if err != nil || !shared.IsNil(result[0]) {
		// Error or not found.
		return result, err
	}

	query = shared.SearchableNormalized(artistName, albumName)
	return e.search(ctx, query, true)
}

func (e AlbumsSearchAction) search(
	ctx context.Context,
	query string,
	useSuggest bool,
) ([10]shared.RemoteAlbum, error) {

	var (
		err      error
		result   [10]shared.RemoteAlbum
		sug      *Album
		sugQuery string
	)

	if useSuggest {
		sug, sugQuery, err = e.suggest(ctx, query)
		if err != nil {
			return result, err
		}
		// If no suggests,
		// try to change suggested query if exists.
		if sug == nil && len(sugQuery) > 0 {
			query = sugQuery
		}
	}

	search, err := e.client.Search(ctx, query, 0, schema.SearchTypeAlbum, false)
	if err != nil {
		return result, err
	}
	if len(search.Result.Albums.Results) == 0 {
		return result, nil
	}

	i := 0
	if sug != nil {
		result[i] = sug
		i = 1
	}

	for i < 10 {
		if i >= len(search.Result.Albums.Results) {
			break
		}

		// Get full album (with duplicates field).
		// In duplicates can be deluxe versions, remixes, etc.
		full, err := e.client.Album(ctx, search.Result.Albums.Results[i].ID, true)
		notFound, err := isNotFoundOrErr(err, full.Result.Title)
		if err != nil {
			return result, err
		}
		if notFound {
			continue
		}

		alb, err := newAlbum(full.Result, e.client)
		if err != nil {
			return result, err
		}
		result[i] = alb
		i++

		for x := range full.Result.Duplicates {
			if i > 4 {
				// Keep for other albums.
				break
			}
			albDup, err := newAlbum(full.Result.Duplicates[x], e.client)
			if err != nil {
				return result, err
			}
			result[i] = albDup
			i++
		}

	}

	return result, err
}

// Suggested, suggested query, error.
func (e AlbumsSearchAction) suggest(ctx context.Context, query string) (*Album, string, error) {
	var (
		sug      *Album
		sugQuery string
	)

	suggests, err := e.client.SearchSuggest(ctx, query)
	if err != nil {
		return nil, "", err
	}
	if suggests.Result.Best.Album != nil {
		// The search prompts have a fucked up structure, so you have to do tricks to get the suggested album ID.
		alb, err := newAlbum(*suggests.Result.Best.Album, e.client)
		if err != nil {
			return sug, sugQuery, err
		}
		sug = alb
	}
	if len(suggests.Result.Best.Text) > 0 {
		// Here is a suggested query that will probably fit.
		sugQuery = suggests.Result.Best.Text
	}

	return sug, sugQuery, err
}

type ArtistsSearchAction struct {
	client *goym.Client
}

func (e ArtistsSearchAction) Search(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	var result [10]shared.RemoteArtist

	query := what.Name()

	search, err := e.client.Search(ctx, query, 0, schema.SearchTypeArtist, false)
	if err != nil {
		return result, err
	}

	for i := range result {
		if i == len(search.Result.Artists.Results) {
			break
		}
		artist, err := newArtist(search.Result.Artists.Results[i], e.client)
		if err != nil {
			return result, err
		}
		result[i] = artist
	}

	return result, err
}

type TracksSearchAction struct {
	client *goym.Client
}

func (e TracksSearchAction) Search(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	var result [10]shared.RemoteTrack

	query := what.Artists()[0].Name() + " " + shared.SearchablePart(what.Name())

	search, err := e.client.Search(ctx, query, 0, schema.SearchTypeTrack, false)
	if err != nil {
		return result, err
	}

	for i := range search.Result.Tracks.Results {
		if i == len(result) {
			break
		}
		if isUgcTrack(search.Result.Tracks.Results[i]) {
			continue
		}
		track, err := newTrack(search.Result.Tracks.Results[i], e.client)
		if err != nil {
			return result, err
		}
		result[i] = track
	}

	return result, err
}
