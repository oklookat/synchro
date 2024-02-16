package zvuk

import (
	"context"

	"github.com/oklookat/gozvuk"
	"github.com/oklookat/gozvuk/schema"
	"github.com/oklookat/synchro/shared"
)

const (
	// You can't keep a playlist empty, so there must be some track in it.
	_tempPlaylistTrackId schema.ID = "132825405"
)

func newPlaylist(account shared.Account, playlist schema.Playlist, client *gozvuk.Client) *Playlist {
	return &Playlist{
		Entity:   newEntity(string(playlist.ID), playlist.Title),
		account:  account,
		playlist: playlist,
		client:   client,
	}
}

type Playlist struct {
	*Entity
	account shared.Account

	playlist schema.Playlist
	client   *gozvuk.Client

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

	result := map[shared.RemoteID]shared.RemoteTrack{}

	for i := range e.cachedTracks {
		// Skip temp track.
		if e.cachedTracks[i].ID == _tempPlaylistTrackId {
			continue
		}
		result[shared.RemoteID(e.cachedTracks[i].ID)] = newTrack(e.cachedTracks[i], e.client)
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (e *Playlist) Rename(ctx context.Context, newName string) error {
	_, err := e.client.RenamePlaylist(ctx, e.playlist.ID, newName)
	if err != nil {
		return err
	}
	e.playlist.Title = newName
	return err
}

func (e *Playlist) SetDescription(ctx context.Context, newDesc string) error {
	return shared.ErrNotImplemented
}

func (e *Playlist) AddTracks(ctx context.Context, ids []shared.RemoteID) error {
	if len(ids) == 0 {
		return nil
	}

	var converted []schema.ID
	for _, id := range ids {
		for _, v := range e.playlist.Tracks {
			if v.ID == schema.ID(id.String()) {
				// Track exists.
				continue
			}
		}
		converted = append(converted, schema.ID(id))
	}

	// 25 items per request.
	idsChunked := shared.ChunkSlice(converted, 25)
	for i := range idsChunked {
		var items []schema.PlaylistItem
		for _, addId := range idsChunked[i] {
			items = append(items, schema.PlaylistItem{
				Type:   schema.PlaylistItemTypeTrack,
				ItemID: addId,
			})
		}
		if len(items) == 0 {
			continue
		}
		_, err := e.client.AddTracksToPlaylist(ctx, e.playlist.ID, items)
		if err != nil {
			return err
		}
		for _, addedId := range idsChunked[i] {
			e.playlist.Tracks = append(e.playlist.Tracks, struct {
				ID schema.ID "json:\"id\""
			}{
				ID: addedId,
			})
		}
	}

	// Remove temp track.
	if ids[0] != shared.RemoteID(_tempPlaylistTrackId) && len(e.playlist.Tracks) > 1 {
		if err := e.RemoveTracks(ctx, []shared.RemoteID{shared.RemoteID(_tempPlaylistTrackId)}); err != nil {
			return err
		}
	}

	e.cachedTracks = nil
	return nil
}

func (e *Playlist) RemoveTracks(ctx context.Context, ids []shared.RemoteID) error {
	var items []schema.PlaylistItem
	deletedIndexes := map[int]bool{}

	keepTempTrack := len(ids) >= len(e.playlist.Tracks)
	if keepTempTrack {
		// Keep one track, because we can't have
		// playlist without tracks.
		if err := e.AddTracks(ctx, []shared.RemoteID{shared.RemoteID(_tempPlaylistTrackId)}); err != nil {
			return err
		}
	}

	for i, tr := range e.playlist.Tracks {
		if tr.ID == _tempPlaylistTrackId && keepTempTrack {
			items = append(items, schema.PlaylistItem{
				Type:   schema.PlaylistItemTypeTrack,
				ItemID: tr.ID,
			})
			continue
		}

		for _, removeID := range ids {
			if tr.ID == schema.ID(removeID) {
				// Track will be removed.
				deletedIndexes[i] = true
				break
			}
		}

		if _, ok := deletedIndexes[i]; ok {
			continue
		}

		// Track not be removed.
		items = append(items, schema.PlaylistItem{
			Type:   schema.PlaylistItemTypeTrack,
			ItemID: tr.ID,
		})
	}

	if len(items) == 0 {
		return nil
	}

	var filteredTracks []struct {
		ID schema.ID "json:\"id\""
	}
	_, err := e.client.UpdataPlaylist(ctx, e.playlist.ID, items, e.playlist.IsPublic, e.playlist.Title)
	if err == nil {
		for i, v := range e.playlist.Tracks {
			if _, ok := deletedIndexes[i]; ok {
				continue
			}
			filteredTracks = append(filteredTracks, struct {
				ID schema.ID "json:\"id\""
			}{v.ID})
		}
		e.playlist.Tracks = filteredTracks
	}

	e.cachedTracks = nil
	return err
}

func (e *Playlist) IsVisible() (bool, error) {
	return e.playlist.IsPublic, nil
}

func (e *Playlist) SetIsVisible(ctx context.Context, val bool) error {
	_, err := e.client.SetPlaylistToPublic(ctx, e.playlist.ID, val)
	return err
}

func (e *Playlist) cacheTracks(ctx context.Context) error {
	if len(e.cachedTracks) > 0 {
		return nil
	}

	chunks := shared.ChunkSlice(e.playlist.Tracks, 25)
	for _, chunk := range chunks {
		var ids []schema.ID
		for _, item := range chunk {
			if item.ID == _tempPlaylistTrackId {
				continue
			}
			ids = append(ids, item.ID)
		}

		if len(ids) == 0 {
			continue
		}

		resp, err := e.client.GetFullTrack(ctx, ids)
		if err != nil {
			return err
		}
		e.cachedTracks = append(e.cachedTracks, resp.Data.GetTracks...)
	}

	return nil
}
