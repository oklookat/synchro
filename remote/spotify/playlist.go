package spotify

import (
	"context"

	"github.com/oklookat/synchro/shared"
	"github.com/zmb3/spotify/v2"
)

func newPlaylist(account shared.Account, playlist spotify.SimplePlaylist, client *spotify.Client) *Playlist {
	return &Playlist{
		Entity:     newEntity(playlist.ID.String(), playlist.Name),
		account:    account,
		playlist:   playlist,
		client:     client,
		snapshotID: playlist.SnapshotID,
	}
}

type Playlist struct {
	*Entity
	account shared.Account

	playlist     spotify.SimplePlaylist
	cachedTracks map[spotify.ID]spotify.FullTrack

	snapshotID string
	client     *spotify.Client
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
	if len(e.cachedTracks) == 0 {
		return nil, nil
	}

	result := map[shared.RemoteID]shared.RemoteTrack{}
	for id, track := range e.cachedTracks {
		trackd := track
		result[shared.RemoteID(id.String())] = newTrack(trackd, e.client)
	}

	return result, nil
}

func (e *Playlist) Rename(ctx context.Context, newName string) error {
	err := e.client.ChangePlaylistName(ctx, e.playlist.ID, newName)
	if err == nil {
		e.playlist.Name = newName
	}
	return err
}

func (e Playlist) SetDescription(ctx context.Context, newDesc string) error {
	return e.client.ChangePlaylistDescription(ctx, e.playlist.ID, newDesc)
}

func (e *Playlist) IsVisible() (bool, error) {
	return e.playlist.IsPublic, nil
}

func (e *Playlist) SetIsVisible(ctx context.Context, val bool) error {
	if err := e.client.ChangePlaylistAccess(ctx, e.playlist.ID, val); err != nil {
		return err
	}
	e.playlist.IsPublic = val
	return nil
}

func (e *Playlist) AddTracks(ctx context.Context, ids []shared.RemoteID) error {
	return e.addRemoveTracks(ctx, ids, true)
}

func (e *Playlist) RemoveTracks(ctx context.Context, ids []shared.RemoteID) error {
	return e.addRemoveTracks(ctx, ids, false)
}

func (e *Playlist) addRemoveTracks(ctx context.Context, ids []shared.RemoteID, add bool) error {
	e.cachedTracks = nil

	var converted []spotify.ID
	for _, id := range ids {
		converted = append(converted, spotify.ID(id))
	}

	// 80 items per request.
	idsChunked := shared.ChunkSlice(converted, 80)
	for i := range idsChunked {
		var snapshotID string
		var err error
		if add {
			snapshotID, err = e.client.AddTracksToPlaylist(ctx, e.playlist.ID, idsChunked[i]...)
		} else {
			snapshotID, err = e.client.RemoveTracksFromPlaylist(ctx, e.playlist.ID, idsChunked[i]...)
		}
		if err != nil {
			return err
		}
		e.snapshotID = snapshotID
	}

	return nil
}

func (e *Playlist) cacheTracks(ctx context.Context) error {
	if len(e.cachedTracks) > 0 {
		return nil
	}

	const limit = 45
	offset := 0

	e.cachedTracks = map[spotify.ID]spotify.FullTrack{}

	for {
		page, err := e.client.GetPlaylistItems(ctx, e.playlist.ID, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return err
		}

		for _, item := range page.Items {
			if item.IsLocal {
				continue
			}
			if item.Track.Track == nil {
				continue
			}
			if item.Track.Track.Type != "track" {
				continue
			}
			track := *item.Track.Track
			e.cachedTracks[track.ID] = *item.Track.Track

		}

		if len(page.Next) == 0 {
			break
		}

		offset += limit
	}

	return nil
}
