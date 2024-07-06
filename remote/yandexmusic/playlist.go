package yandexmusic

import (
	"context"

	"github.com/oklookat/goym"
	"github.com/oklookat/goym/schema"
	"github.com/oklookat/synchro/shared"
)

func newPlaylist(account shared.Account, playlist schema.Playlist, client *goym.Client) *Playlist {
	return &Playlist{
		Entity:   newEntity(playlist.Kind.String(), playlist.Title),
		account:  account,
		playlist: playlist,
		client:   client,
	}
}

type Playlist struct {
	*Entity
	account shared.Account

	playlist   schema.Playlist
	client     *goym.Client
	trackItems []schema.TrackItem
}

func (e Playlist) FromAccount() shared.Account {
	return e.account
}

func (e Playlist) Description() *string {
	return &e.playlist.Description
}

func (e *Playlist) Tracks(ctx context.Context) ([]shared.RemoteTrack, error) {
	if err := e.cacheTracks(ctx); err != nil {
		return nil, err
	}

	if len(e.trackItems) == 0 {
		return nil, nil
	}

	result := []shared.RemoteTrack{}

	for _, item := range e.trackItems {
		track, err := newTrack(item.Track, e.client)
		if err != nil {
			return nil, err
		}
		result = append(result, track)
	}

	return result, nil
}

func (e *Playlist) Rename(ctx context.Context, newName string) error {
	pl, err := e.client.RenamePlaylist(ctx, e.playlist.Kind, newName)
	if err != nil {
		return err
	}
	e.playlist = pl.Result
	return err
}

func (e *Playlist) SetDescription(ctx context.Context, newDesc string) error {
	pl, err := e.client.SetPlaylistDescription(ctx, e.playlist.Kind, newDesc)
	if err != nil {
		return err
	}
	e.playlist = pl.Result
	return err
}

func (e *Playlist) AddTracks(ctx context.Context, ids []shared.RemoteID) error {
	if len(ids) == 0 {
		return nil
	}

	var converted []schema.ID
	for _, id := range ids {
		converted = append(converted, schema.ID(id))
	}

	var toAdd []schema.Track

	// 25 items per request.
	idsChunked := shared.ChunkSlice(converted, 25)
	for i := range idsChunked {
		tracks, err := e.client.Tracks(ctx, idsChunked[i])
		if err != nil {
			return err
		}
		toAdd = append(toAdd, tracks.Result...)
		pl, err := e.client.AddToPlaylist(ctx, e.playlist, toAdd)
		if err != nil {
			return err
		}
		e.playlist = pl.Result
	}

	e.trackItems = nil
	return nil
}

func (e *Playlist) RemoveTracks(ctx context.Context, ids []shared.RemoteID) error {
	if len(ids) == 0 {
		return nil
	}

	if err := e.cacheTracks(ctx); err != nil {
		return err
	}

	converted := make([]schema.ID, len(ids))
	for i := range converted {
		converted[i] = schema.ID(ids[i])
	}

	// 25 items per request.
	idsChunked := shared.ChunkSlice(converted, 25)
	for i := range idsChunked {
		resp, err := e.client.DeleteTracksFromPlaylist(ctx, e.playlist, idsChunked[i])
		if err != nil {
			return err
		}
		e.playlist = resp.Result
	}

	e.trackItems = nil
	return nil
}

func (e *Playlist) IsVisible() (bool, error) {
	return e.playlist.Visibility == schema.VisibilityPublic, nil
}

func (e *Playlist) SetIsVisible(ctx context.Context, val bool) error {
	vis := schema.VisibilityPublic
	if !val {
		vis = schema.VisibilityPrivate
	}

	pl, err := e.client.SetPlaylistVisibility(ctx, e.playlist.Kind, vis)
	if err != nil {
		return err
	}

	e.playlist = pl.Result
	e.trackItems = nil
	return err
}

func (e *Playlist) cacheTracks(ctx context.Context) error {
	if len(e.trackItems) > 0 {
		return nil
	}

	pl, err := e.client.MyPlaylist(ctx, e.playlist.Kind)
	if err != nil {
		return err
	}

	e.trackItems = []schema.TrackItem{}
	e.trackItems = append(e.trackItems, pl.Result.Tracks...)

	return err
}
