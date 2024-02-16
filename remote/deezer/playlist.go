package deezer

import (
	"context"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/schema"
	"github.com/oklookat/synchro/shared"
)

func newPlaylist(
	ctx context.Context,
	cl *deezus.Client,
	acc shared.Account,
	plId schema.ID,
) (*Playlist, error) {

	resp, err := cl.Playlist(ctx, plId)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &Playlist{
		Entity:   newEntity(resp.Playlist.ID.String(), resp.Playlist.Title),
		client:   cl,
		account:  acc,
		playlist: resp.Playlist,
	}, err
}

type Playlist struct {
	*Entity

	client   *deezus.Client
	account  shared.Account
	playlist schema.Playlist

	cachedTracks map[shared.RemoteID]shared.RemoteTrack
}

func (e Playlist) FromAccount() shared.Account {
	return e.account
}

func (e Playlist) Description() *string {
	return &e.playlist.Description
}

func (e *Playlist) Tracks(ctx context.Context) (map[shared.RemoteID]shared.RemoteTrack, error) {
	if err := e.cacheTracks(ctx); err != nil {
		return nil, err
	}
	return e.cachedTracks, nil
}

func (e *Playlist) Rename(ctx context.Context, newName string) error {
	_, err := e.client.UpdatePlaylist(ctx, e.playlist.ID, &newName, nil, nil)
	if err == nil {
		e.playlist.Title = newName
	}
	return err
}

func (e Playlist) SetDescription(ctx context.Context, newDesc string) error {
	_, err := e.client.UpdatePlaylist(ctx, e.playlist.ID, nil, &newDesc, nil)
	if err == nil {
		e.playlist.Description = newDesc
	}
	return err
}

func (e *Playlist) IsVisible() (bool, error) {
	return e.playlist.Public, nil
}

func (e *Playlist) SetIsVisible(ctx context.Context, val bool) error {
	_, err := e.client.UpdatePlaylist(ctx, e.playlist.ID, nil, nil, &val)
	if err == nil {
		e.playlist.Public = val
	}
	return err
}

func (e *Playlist) AddTracks(ctx context.Context, ids []shared.RemoteID) error {
	return e.addRemoveTracks(ctx, ids, true)
}

func (e *Playlist) RemoveTracks(ctx context.Context, ids []shared.RemoteID) error {
	return e.addRemoveTracks(ctx, ids, false)
}

func (e *Playlist) addRemoveTracks(ctx context.Context, ids []shared.RemoteID, add bool) error {
	e.cachedTracks = nil

	var converted []schema.ID
	for _, id := range ids {
		conv, err := remoteToSchemaID(id)
		if err != nil {
			return err
		}
		converted = append(converted, conv)
	}

	idsChunked := shared.ChunkSlice(converted, 30)
	for _, chunk := range idsChunked {
		if add {
			_, err := e.client.AddTracksToPlaylist(ctx, e.playlist.ID, chunk)
			if err != nil {
				return err
			}
			continue
		}
		_, err := e.client.RemoveTracksFromPlaylist(ctx, e.playlist.ID, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Playlist) cacheTracks(ctx context.Context) error {
	if len(e.cachedTracks) > 0 {
		return nil
	}

	const limit = 30
	offset := 0

	e.cachedTracks = map[shared.RemoteID]shared.RemoteTrack{}

	for {
		resp, err := e.client.PlaylistTracks(ctx, e.playlist.ID, offset, limit)
		if err != nil {
			return err
		}

		for _, item := range resp.Data {
			conv, err := newTrack(ctx, e.client, item.ID)
			if err != nil {
				return err
			}
			e.cachedTracks[shared.RemoteID(item.ID.String())] = conv

		}

		if len(resp.Data) == 0 || resp.Next == nil || len(*resp.Next) == 0 {
			break
		}

		offset += limit
	}

	return nil
}
