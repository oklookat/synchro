package vkmusic

import (
	"context"

	"github.com/oklookat/govkm"
	"github.com/oklookat/govkm/schema"
	"github.com/oklookat/synchro/shared"
)

func newPlaylist(
	account shared.Account,
	playlist schema.Playlist,
	client *govkm.Client,
) *Playlist {

	if account == nil || client == nil {
		return nil
	}

	return &Playlist{
		Entity:   newEntity(playlist.APIID.String(), playlist.Name),
		account:  account,
		playlist: playlist,
		client:   client,
	}
}

type Playlist struct {
	*Entity
	account shared.Account

	playlist schema.Playlist

	client       *govkm.Client
	cachedTracks []schema.Track
}

func (e Playlist) FromAccount() shared.Account {
	return e.account
}

func (e Playlist) Description() *string {
	return nil
}

func (e *Playlist) Tracks(ctx context.Context) (map[shared.RemoteID]shared.RemoteTrack, error) {
	if err := e.cacheTracks(ctx); err != nil {
		return nil, err
	}
	if len(e.cachedTracks) == 0 {
		return nil, nil
	}

	result := map[shared.RemoteID]shared.RemoteTrack{}
	for i := range e.cachedTracks {
		track, err := newTrack(e.cachedTracks[i], e.client)
		if err != nil {
			return nil, err
		}
		result[shared.RemoteID(e.cachedTracks[i].APIID)] = track
	}

	return result, nil
}

func (e *Playlist) Rename(ctx context.Context, newName string) error {
	_, err := e.client.RenamePlaylist(ctx, e.playlist.APIID, newName)
	return err
}

func (e Playlist) SetDescription(ctx context.Context, newDesc string) error {
	return shared.ErrNotImplemented
}

func (e *Playlist) IsVisible() (bool, error) {
	return false, shared.ErrNotImplemented
}

func (e *Playlist) SetIsVisible(ctx context.Context, val bool) error {
	return shared.ErrNotImplemented
}

func (e *Playlist) AddTracks(ctx context.Context, ids []shared.RemoteID) error {
	if len(ids) == 0 {
		return nil
	}

	if err := e.cacheTracks(ctx); err != nil {
		return err
	}

	var converted []schema.ID
	for _, id := range ids {
		for _, v := range e.cachedTracks {
			if v.APIID == schema.ID(id.String()) {
				// Track exists.
				continue
			}
		}
		converted = append(converted, schema.ID(id))
	}

	for _, id := range converted {
		pl, err := e.client.AddTrackToPlaylist(ctx, e.playlist.APIID, id)
		if err != nil {
			return err
		}
		if pl.Data.Playlist != nil {
			e.playlist = *pl.Data.Playlist
		}
	}

	e.cachedTracks = nil
	return nil
}

func (e *Playlist) RemoveTracks(ctx context.Context, ids []shared.RemoteID) error {
	if err := e.cacheTracks(ctx); err != nil {
		return err
	}

	var trackIds []schema.ID
	deletedIndexes := map[int]bool{}

	for i, tr := range e.cachedTracks {
		for _, removeID := range ids {
			if tr.APIID == schema.ID(removeID) {
				// Track will be removed.
				deletedIndexes[i] = true
				break
			}
		}
		if _, ok := deletedIndexes[i]; ok {
			continue
		}

		trackIds = append(trackIds, tr.APIID)
	}

	if len(trackIds) == 0 {
		return nil
	}

	resp, err := e.client.EditPlaylist(ctx,
		e.playlist.Name,
		e.playlist.APIID, trackIds)
	if err != nil {
		return err
	}
	if resp.Data.Playlist != nil {
		e.playlist = *resp.Data.Playlist
	}

	e.cachedTracks = nil
	return err
}

func (e *Playlist) cacheTracks(ctx context.Context) error {
	if len(e.cachedTracks) > 0 {
		return nil
	}

	var tracks []schema.Track

	offset := 0
	for {
		data, err := e.client.PlaylistTracks(ctx, e.playlist.APIID, 30, offset)
		if err != nil {
			return err
		}

		if len(data.Data.Tracks) == 0 {
			break
		}
		tracks = append(tracks, data.Data.Tracks...)
		offset += len(data.Data.Tracks)
	}

	e.cachedTracks = tracks

	return nil
}
