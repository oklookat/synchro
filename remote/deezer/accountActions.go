package deezer

import (
	"context"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/schema"
	"github.com/oklookat/synchro/shared"
)

func newAccountActions(account shared.Account) (*AccountActions, error) {
	client, err := getClient(account)
	return &AccountActions{
		account: account,
		client:  client,
	}, err
}

type AccountActions struct {
	account shared.Account
	client  *deezus.Client
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
	client *deezus.Client
}

func (e LikedAlbumsActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	albums := []shared.RemoteEntity{}

	offset := 0
	const limit = 60

	for {
		albumsd, err := e.client.UserMeAlbums(ctx, offset, limit)
		if err != nil {
			if isNotFound(err) {
				break
			}
			return nil, err
		}

		if len(albumsd.Data) == 0 {
			break
		}

		for _, al := range albumsd.Data {
			conv, err := newAlbum(ctx, e.client, al.ID)
			if err != nil {
				return nil, err
			}
			albums = append(albums, conv)
		}

		offset += limit
	}

	return albums, nil
}

func (e LikedAlbumsActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedAlbumsActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedAlbumsActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	return addRemove(ctx, e.client, ids, like, _entityTypeAlbum)
}

type LikedArtistsActions struct {
	client *deezus.Client
}

func (e LikedArtistsActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	result := []shared.RemoteEntity{}

	const limit = 60
	offset := 0

	for {
		resp, err := e.client.UserMeArtists(ctx, offset, limit)
		if err != nil {
			if isNotFound(err) {
				err = nil
			}
			return nil, err
		}

		for i := range resp.Data {
			result = append(result, newArtist(e.client, resp.Data[i]))
		}

		if len(resp.Data) == 0 || resp.Next == nil || len(*resp.Next) == 0 {
			break
		}

		offset += limit
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
	return addRemove(ctx, e.client, ids, like, _entityTypeArtist)
}

type LikedTracksActions struct {
	client *deezus.Client
}

func (e LikedTracksActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	result := []shared.RemoteEntity{}

	const limit = 60
	offset := 0

	for {
		resp, err := e.client.UserMeTracks(ctx, offset, limit)
		if err != nil {
			if isNotFound(err) {
				err = nil
			}
			return nil, err
		}

		for i := range resp.Data {
			conv, err := newTrack(ctx, e.client, resp.Data[i].ID)
			if err != nil {
				return nil, err
			}
			result = append(result, conv)
		}

		if len(resp.Data) == 0 || resp.Next == nil || len(*resp.Next) == 0 {
			break
		}

		offset += limit
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
	return addRemove(ctx, e.client, ids, like, _entityTypeTrack)
}

type PlaylistActions struct {
	account shared.Account
	client  *deezus.Client
}

func (e *PlaylistActions) MyPlaylists(ctx context.Context) ([]shared.RemotePlaylist, error) {
	result := []shared.RemotePlaylist{}

	const limit = 60
	offset := 0

	for {
		resp, err := e.client.UserMePlaylists(ctx, offset, limit)
		if err != nil {
			if isNotFound(err) {
				err = nil
			}
			return nil, err
		}

		for i := range resp.Data {
			if resp.Data[i].Collaborative || resp.Data[i].IsLovedTrack {
				continue
			}
			if resp.Data[i].Creator != nil && resp.Data[i].Creator.ID != e.client.UserID {
				continue
			}
			conv, err := newPlaylist(ctx, e.client, e.account, resp.Data[i].ID)
			if err != nil {
				return nil, err
			}
			result = append(result, conv)
		}

		if len(resp.Data) == 0 || resp.Next == nil || len(*resp.Next) == 0 {
			break
		}

		offset += limit
	}

	return result, nil
}

func (e PlaylistActions) Create(ctx context.Context, name string, isVisible bool, description *string) (shared.RemotePlaylist, error) {
	ideed, err := e.client.CreatePlaylist(ctx, name)
	if err != nil {
		return nil, err
	}
	if _, err = e.client.UpdatePlaylist(ctx, ideed.ID, nil, description, &isVisible); err != nil {
		return nil, err
	}
	return newPlaylist(ctx, e.client, e.account, ideed.ID)
}

func (e PlaylistActions) Delete(ctx context.Context, entities []shared.RemoteID) error {
	for _, id := range entities {
		conv, err := remoteToSchemaID(id)
		if err != nil {
			return err
		}
		if _, err := e.client.DeletePlaylist(ctx, conv); err != nil {
			return err
		}
	}
	return nil
}

func (e PlaylistActions) Playlist(ctx context.Context, id shared.RemoteID) (shared.RemotePlaylist, error) {
	conv, err := remoteToSchemaID(id)
	if err != nil {
		return nil, err
	}
	return newPlaylist(ctx, e.client, e.account, conv)
}

type entityType int

const (
	_entityTypeAlbum entityType = iota
	_entityTypeArtist
	_entityTypeTrack
)

func addRemove(
	ctx context.Context,
	cl *deezus.Client,
	ids []shared.RemoteID,
	add bool,
	ent entityType,
) error {

	converted := make([]schema.ID, len(ids))
	for i := range converted {
		conv, err := remoteToSchemaID(ids[i])
		if err != nil {
			return err
		}
		converted[i] = conv
	}

	remStatic := func(rem func(ctx context.Context, id schema.ID) (*schema.BoolResponse, error)) error {
		for _, id := range converted {
			if _, err := rem(ctx, id); err != nil {
				return err
			}
		}
		return nil
	}
	addStatic := func(add func(ctx context.Context, ids []schema.ID) (*schema.BoolResponse, error)) error {
		idsChunked := shared.ChunkSlice(converted, 25)
		for _, chunk := range idsChunked {
			if _, err := add(ctx, chunk); err != nil {
				return err
			}
		}
		return nil
	}

	switch ent {
	case _entityTypeAlbum:
		if add {
			return addStatic(cl.AddAlbums)
		}
		return remStatic(cl.RemoveAlbum)
	case _entityTypeArtist:
		if add {
			return addStatic(cl.AddArtists)
		}
		return remStatic(cl.RemoveArtist)
	case _entityTypeTrack:
		if add {
			return addStatic(cl.AddTracks)
		}
		return remStatic(cl.RemoveTrack)
	}

	return nil
}
