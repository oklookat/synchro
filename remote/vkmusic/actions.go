package vkmusic

import (
	"context"

	"github.com/oklookat/govkm"
	"github.com/oklookat/govkm/schema"
	"github.com/oklookat/synchro/shared"
)

func newActions(client *govkm.Client) *Actions {
	if client == nil {
		return nil
	}
	return &Actions{
		client: client,
	}
}

type Actions struct {
	client *govkm.Client
}

func (e Actions) Album(ctx context.Context, id shared.RemoteID) (shared.RemoteAlbum, error) {
	resp, err := e.client.Album(ctx, schema.ID(id))
	if err != nil {
		return nil, err
	}
	if resp.Data.Album == nil {
		return nil, nil
	}
	return newAlbum(resp.Data.Album, e.client), err
}

func (e Actions) Artist(ctx context.Context, id shared.RemoteID) (shared.RemoteArtist, error) {
	resp, err := e.client.Artist(ctx, schema.ID(id))
	if err != nil {
		return nil, err
	}
	if resp.Data.Artist == nil {
		return nil, nil
	}
	return newArtist(resp.Data.Artist.SimpleArtist, e.client), err
}

func (e Actions) Track(ctx context.Context, id shared.RemoteID) (shared.RemoteTrack, error) {
	resp, err := e.client.Track(ctx, schema.ID(id))
	if err != nil {
		return nil, err
	}
	if resp.Data.Track == nil {
		return nil, nil
	}
	return newTrack(*resp.Data.Track, e.client)
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
	client *govkm.Client
}

func (e AlbumsSearchAction) Search(ctx context.Context, what shared.RemoteAlbum) ([10]shared.RemoteAlbum, error) {
	var result [10]shared.RemoteAlbum

	artistName := what.Artists()[0].Name()
	albumName := what.Name()
	query := artistName + " " + shared.SearchablePart(albumName)

	resp, err := e.client.SearchAlbum(ctx, query, 10, 0)
	if err != nil {
		return result, err
	}

	for i := range result {
		if i == len(resp.Data.Albums) {
			break
		}
		result[i] = newAlbum(&resp.Data.Albums[i], e.client)
	}

	return result, err
}

type ArtistsSearchAction struct {
	client *govkm.Client
}

func (e ArtistsSearchAction) Search(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	var result [10]shared.RemoteArtist

	query := what.Name()

	resp, err := e.client.SearchArtist(ctx, query, 11, 0)
	if err != nil {
		return result, err
	}

	for i := range result {
		if i == len(resp.Data.Artists) {
			break
		}
		result[i] = newArtist(resp.Data.Artists[i].SimpleArtist, e.client)
	}

	return result, err
}

type TracksSearchAction struct {
	client *govkm.Client
}

func (e TracksSearchAction) Search(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	var result [10]shared.RemoteTrack

	artistName := what.Artists()[0].Name()
	trackName := what.Name()
	query := artistName + " " + shared.SearchablePart(trackName)

	resp, err := e.client.SearchTrack(ctx, query, 11, 0)
	if err != nil {
		return result, err
	}

	for i := range resp.Data.Tracks {
		if i == len(result) {
			break
		}
		// if isUgcTrack(resp.Data.Tracks[i]) {
		// 	continue
		// }
		track, err := newTrack(resp.Data.Tracks[i], e.client)
		if err != nil {
			return result, err
		}
		result[i] = track
	}

	return result, err
}
