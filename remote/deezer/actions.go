package deezer

import (
	"context"

	"github.com/oklookat/deezus"
	"github.com/oklookat/synchro/shared"
)

func newActions(client *deezus.Client) *Actions {
	if client == nil {
		return nil
	}
	return &Actions{
		client: client,
	}
}

type Actions struct {
	client *deezus.Client
}

func (e Actions) Album(ctx context.Context, id shared.RemoteID) (shared.RemoteAlbum, error) {
	conv, err := remoteToSchemaID(id)
	if err != nil {
		return nil, err
	}
	return newAlbum(ctx, e.client, conv)
}

func (e Actions) Artist(ctx context.Context, id shared.RemoteID) (shared.RemoteArtist, error) {
	conv, err := remoteToSchemaID(id)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Artist(ctx, conv)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return newArtist(e.client, resp.SimpleArtist), err
}

func (e Actions) Track(ctx context.Context, id shared.RemoteID) (shared.RemoteTrack, error) {
	conv, err := remoteToSchemaID(id)
	if err != nil {
		return nil, err
	}
	return newTrack(ctx, e.client, conv)
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
	client *deezus.Client
}

func (e AlbumsSearchAction) Search(ctx context.Context, what shared.RemoteAlbum) ([10]shared.RemoteAlbum, error) {
	var result [10]shared.RemoteAlbum

	query := what.Artists()[0].Name() + " " + shared.SearchablePart(what.Name())

	resp, err := e.client.SearchAlbums(ctx, query, "", false, 0, 11)
	if err != nil {
		if isNotFound(err) {
			err = nil
		}
		return result, err
	}

	for i := range result {
		if i == len(resp.Data) {
			break
		}
		conv, err := newAlbum(ctx, e.client, resp.Data[i].ID)
		if err != nil {
			return result, err
		}
		result[i] = conv
	}

	return result, err
}

type ArtistsSearchAction struct {
	client *deezus.Client
}

func (e ArtistsSearchAction) Search(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	var result [10]shared.RemoteArtist

	query := what.Name()

	resp, err := e.client.SearchArtists(ctx, query, "", false, 0, 11)
	if err != nil {
		if isNotFound(err) {
			err = nil
		}
		return result, err
	}

	for i := range result {
		if i == len(resp.Data) {
			break
		}
		result[i] = newArtist(e.client, resp.Data[i])
	}

	return result, err
}

type TracksSearchAction struct {
	client *deezus.Client
}

func (e TracksSearchAction) Search(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	var result [10]shared.RemoteTrack

	query := what.Artists()[0].Name() + " " + shared.SearchablePart(what.Name())

	resp, err := e.client.SearchTracks(ctx, query, "", false, 0, 11)
	if err != nil {
		if isNotFound(err) {
			err = nil
		}
		return result, err
	}

	for i := range result {
		if i == len(resp.Data) {
			break
		}
		conv, err := newTrack(ctx, e.client, resp.Data[i].ID)
		if err != nil {
			return result, err
		}
		result[i] = conv
	}

	return result, err
}
