package spotify

import (
	"context"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
	"github.com/zmb3/spotify/v2"
)

func newActions(client *spotify.Client) *Actions {
	if client == nil {
		return nil
	}
	return &Actions{
		client: client,
	}
}

type Actions struct {
	client *spotify.Client
}

func (e Actions) Album(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceAlbum, error) {
	album, err := e.client.GetAlbum(ctx, spotify.ID(id))
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return newAlbum(album, e.client), err
}

func (e Actions) Artist(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceArtist, error) {
	artist, err := e.client.GetArtist(ctx, spotify.ID(id))
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return newArtist(artist.SimpleArtist, e.client), err
}

func (e Actions) Track(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceTrack, error) {
	track, err := e.client.GetTrack(ctx, spotify.ID(id))
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if track == nil {
		return nil, nil
	}
	return newTrack(*track, e.client), err
}

func (e Actions) SearchAlbums(ctx context.Context, what streaming.ServiceAlbum) ([10]streaming.ServiceAlbum, error) {
	action := &AlbumsSearchAction{e.client}
	return action.Search(ctx, what)
}

func (e Actions) SearchArtists(ctx context.Context, what streaming.ServiceArtist) ([10]streaming.ServiceArtist, error) {
	action := &ArtistsSearchAction{e.client}
	return action.Search(ctx, what)
}

func (e Actions) SearchTracks(ctx context.Context, what streaming.ServiceTrack) ([10]streaming.ServiceTrack, error) {
	action := &TracksSearchAction{e.client}
	return action.Search(ctx, what)
}

type AlbumsSearchAction struct {
	client *spotify.Client
}

func (e AlbumsSearchAction) Search(ctx context.Context, what streaming.ServiceAlbum) ([10]streaming.ServiceAlbum, error) {
	var result [10]streaming.ServiceAlbum

	query := what.Artists()[0].Name() + " " + shared.SearchablePart(what.Name())

	search, err := pleaseSearch(ctx, e.client, query, spotify.SearchTypeAlbum, spotify.Limit(10), spotify.Offset(0), _market)
	if err != nil {
		return result, err
	}
	if search.Albums == nil || len(search.Albums.Albums) == 0 {
		return result, nil
	}

	var albumsIds []spotify.ID
	for i := range search.Albums.Albums {
		if i == len(result) || i == len(search.Albums.Albums) {
			break
		}
		albumsIds = append(albumsIds, search.Albums.Albums[i].ID)
	}

	fullAlbums, err := e.client.GetAlbums(ctx, albumsIds[:], _market)
	if err != nil {
		return result, err
	}
	for i := range result {
		if i == len(fullAlbums) {
			break
		}
		if fullAlbums[i] == nil {
			break
		}
		result[i] = newAlbum(fullAlbums[i], e.client)
	}

	return result, err
}

type ArtistsSearchAction struct {
	client *spotify.Client
}

func (e ArtistsSearchAction) Search(ctx context.Context, what streaming.ServiceArtist) ([10]streaming.ServiceArtist, error) {
	var result [10]streaming.ServiceArtist

	query := what.Name()

	search, err := pleaseSearch(ctx, e.client, query, spotify.SearchTypeArtist,
		spotify.Limit(10), spotify.Offset(0), _market)
	if err != nil {
		return result, err
	}
	if search.Artists == nil || len(search.Artists.Artists) == 0 {
		return result, nil
	}

	for i := range result {
		if i == len(search.Artists.Artists) {
			break
		}
		result[i] = newArtist(search.Artists.Artists[i].SimpleArtist, e.client)
	}

	return result, err
}

type TracksSearchAction struct {
	client *spotify.Client
}

func (e TracksSearchAction) Search(ctx context.Context, what streaming.ServiceTrack) ([10]streaming.ServiceTrack, error) {
	var result [10]streaming.ServiceTrack

	query := what.Artists()[0].Name() + " " + what.Name()

	search, err := pleaseSearch(ctx, e.client, query, spotify.SearchTypeTrack, spotify.Limit(10), spotify.Offset(0), _market)
	if err != nil {
		if isNotFound(err) {
			return result, nil
		}
		return result, err
	}
	if search.Tracks == nil || len(search.Tracks.Tracks) == 0 {
		return result, nil
	}

	for i := range result {
		if i == len(search.Tracks.Tracks) {
			break
		}
		result[i] = newTrack(search.Tracks.Tracks[i], e.client)
	}

	return result, err
}
