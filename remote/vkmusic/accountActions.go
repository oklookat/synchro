package vkmusic

import (
	"context"

	"github.com/oklookat/govkm"
	"github.com/oklookat/govkm/schema"
	"github.com/oklookat/synchro/shared"
)

func newAccountActions(account shared.Account) (*AccountActions, error) {
	client, err := getClient(account)
	if err != nil {
		return nil, err
	}
	return &AccountActions{
		account: account,
		client:  client,
	}, err
}

type AccountActions struct {
	account shared.Account
	client  *govkm.Client
}

func (e AccountActions) LikedAlbums() shared.LikedActions {
	return &LikedAlbumsActions{client: e.client}
}

func (e AccountActions) LikedArtists() shared.LikedActions {
	return &LikedArtistsActions{client: e.client}
}

func (e AccountActions) LikedTracks() shared.LikedActions {
	return &LikedTracksActions{client: e.client}
}

func (e AccountActions) Playlist() shared.PlaylistActions {
	return &PlaylistActions{
		account: e.account,
		client:  e.client,
	}
}

type LikedAlbumsActions struct {
	client *govkm.Client
}

func (e LikedAlbumsActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	var albums []schema.Album

	const limit = 30
	offset := 0

	for {
		data, err := e.client.LikedAlbums(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		if len(data.Data.Albums) == 0 {
			break
		}
		albums = append(albums, data.Data.Albums...)
		offset += limit
	}

	result := []shared.RemoteEntity{}
	for i := range albums {
		result = append(result, newAlbum(albums[i], e.client))
	}

	return result, nil
}

func (e LikedAlbumsActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedAlbumsActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedAlbumsActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	for _, id := range ids {
		if like {
			if _, err := e.client.LikeAlbum(ctx, schema.ID(id)); err != nil {
				return err
			}
			continue
		}
		if _, err := e.client.UnlikeAlbum(ctx, schema.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

type LikedArtistsActions struct {
	client *govkm.Client
}

func (e LikedArtistsActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	var artists []schema.Artist

	const limit = 30
	offset := 0

	for {
		data, err := e.client.LikedArtists(ctx, limit, offset)
		if err != nil {
			return nil, err
		}
		if len(data.Data.Artists) == 0 {
			break
		}
		artists = append(artists, data.Data.Artists...)
		offset += limit
	}

	result := []shared.RemoteEntity{}
	for i := range artists {
		result = append(result, newArtist(artists[i].SimpleArtist, e.client))
	}

	return result, nil
}

func (e LikedArtistsActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedArtistsActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedArtistsActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	for _, id := range ids {
		if like {
			if _, err := e.client.LikeArtist(ctx, schema.ID(id)); err != nil {
				return err
			}
			continue
		}
		if _, err := e.client.UnlikeArtist(ctx, schema.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

type LikedTracksActions struct {
	client *govkm.Client
}

func (e LikedTracksActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	likesPl, err := e.client.LikedTracks(ctx)
	if err != nil {
		return nil, err
	}

	var tracks []schema.Track

	const limit = 30
	offset := 0

	for {
		data, err := e.client.PlaylistTracks(ctx, likesPl.APIID, limit, offset)
		if err != nil {
			return nil, err
		}
		if len(data.Data.Tracks) == 0 {
			break
		}
		for i := range data.Data.Tracks {
			if isUgcTrack(data.Data.Tracks[i]) {
				continue
			}
			tracks = append(tracks, data.Data.Tracks[i])
		}
		offset += limit
	}

	result := []shared.RemoteEntity{}
	for i := range tracks {
		track, err := newTrack(tracks[i], e.client)
		if err != nil {
			return nil, err
		}
		result = append(result, track)
	}

	return result, nil
}

func (e LikedTracksActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedTracksActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedTracksActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	for _, id := range ids {
		if like {
			if _, err := e.client.LikeTrack(ctx, schema.ID(id)); err != nil {
				return err
			}
			continue
		}
		if _, err := e.client.UnlikeTrack(ctx, schema.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

type PlaylistActions struct {
	account shared.Account
	client  *govkm.Client
}

func (e *PlaylistActions) MyPlaylists(ctx context.Context) ([]shared.RemotePlaylist, error) {
	var playlists []schema.Playlist

	const limit = 30
	offset := 0

	for {
		data, err := e.client.UserPlaylists(ctx, limit, offset)
		if err != nil {
			return nil, err
		}

		if len(data.Data.Playlists) == 0 {
			break
		}

		for i, p := range data.Data.Playlists {
			if p.Owner.APIID != e.client.CurrentUserId || p.IsDefault || p.IsDownloads || p.IsFavorite {
				// Not user playlist or default playlist.
				continue
			}
			playlists = append(playlists, data.Data.Playlists[i])
		}

		offset += limit
	}

	result := []shared.RemotePlaylist{}
	for i := range playlists {
		result = append(result, newPlaylist(e.account, playlists[i], e.client))
	}

	return result, nil
}

func (e PlaylistActions) Create(ctx context.Context, name string, isVisible bool, description *string) (shared.RemotePlaylist, error) {
	resp, err := e.client.CreatePlaylist(ctx, name)
	if err != nil {
		return nil, err
	}
	if resp.Data.Playlist == nil {
		return nil, errNilPlaylist
	}
	return newPlaylist(e.account, *resp.Data.Playlist, e.client), nil
}

func (e PlaylistActions) Delete(ctx context.Context, entities []shared.RemoteID) error {
	for _, id := range entities {
		if _, err := e.client.DeletePlaylist(ctx, schema.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

func (e PlaylistActions) Playlist(ctx context.Context, id shared.RemoteID) (shared.RemotePlaylist, error) {
	resp, err := e.client.Playlist(ctx, schema.ID(id))
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if resp.Data.Playlist == nil {
		return nil, errNilPlaylist
	}
	return newPlaylist(e.account, *resp.Data.Playlist, e.client), err
}
