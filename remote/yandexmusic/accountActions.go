package yandexmusic

import (
	"context"

	"github.com/oklookat/goym"
	"github.com/oklookat/goym/schema"
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
	client  *goym.Client
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
	client *goym.Client
}

func (e LikedAlbumsActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	resp, err := e.client.LikedAlbums(ctx)
	if err != nil {
		return nil, err
	}
	if len(resp.Result) == 0 {
		return nil, nil
	}

	ids := make([]schema.ID, len(resp.Result))
	for i := range ids {
		ids[i] = resp.Result[i].ID
	}

	result := map[shared.RemoteID]shared.RemoteEntity{}

	// 30 items per request.
	idsChunked := shared.ChunkSlice(ids, 30)
	for i := range idsChunked {
		alb, err := e.client.Albums(ctx, idsChunked[i])
		if err != nil {
			return nil, err
		}
		for x := range alb.Result {
			albWrap, err := newAlbum(alb.Result[x], e.client)
			if err != nil {
				return nil, err
			}
			result[shared.RemoteID(alb.Result[x].ID.String())] = albWrap
		}
	}

	return result, err
}

func (e LikedAlbumsActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedAlbumsActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedAlbumsActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	return likeUnlike(ctx, ids, like, e.client.LikeAlbums, e.client.UnlikeAlbums)
}

type LikedArtistsActions struct {
	client *goym.Client
}

func (e LikedArtistsActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	resp, err := e.client.LikedArtists(ctx)
	if err != nil {
		return nil, err
	}

	result := map[shared.RemoteID]shared.RemoteEntity{}
	for i := range resp.Result {
		art, err := newArtist(resp.Result[i], e.client)
		if err != nil {
			return nil, err
		}
		result[shared.RemoteID(resp.Result[i].ID.String())] = art
	}

	return result, err
}

func (e LikedArtistsActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedArtistsActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedArtistsActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	return likeUnlike(ctx, ids, like, e.client.LikeArtists, e.client.UnlikeArtists)
}

type LikedTracksActions struct {
	client *goym.Client
}

func (e LikedTracksActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	resp, err := e.client.LikedTracks(ctx)
	if err != nil {
		return nil, err
	}
	if len(resp.Result.Library.Tracks) == 0 {
		return nil, err
	}

	lib := resp.Result.Library.Tracks
	ids := make([]schema.ID, len(lib))
	for i := range ids {
		ids[i] = lib[i].ID
	}

	result := map[shared.RemoteID]shared.RemoteEntity{}

	// 30 items per request.
	idsChunked := shared.ChunkSlice(ids, 30)
	for i := range idsChunked {
		track, err := e.client.Tracks(ctx, idsChunked[i])
		if err != nil {
			return nil, err
		}
		for x := range track.Result {
			if isUgcTrack(track.Result[x]) {
				continue
			}
			trackWrap, err := newTrack(track.Result[x], e.client)
			if err != nil {
				return nil, err
			}
			result[shared.RemoteID(track.Result[x].ID.String())] = trackWrap
		}
	}

	return result, err
}

func (e LikedTracksActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedTracksActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedTracksActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	return likeUnlike(ctx, ids, like, e.client.LikeTracks, e.client.UnlikeTracks)
}

func likeUnlike(
	ctx context.Context,
	ids []shared.RemoteID,
	like bool,
	liker func(ctx context.Context, ids []schema.ID) (schema.Response[string], error),
	unliker func(ctx context.Context, ids []schema.ID) (schema.Response[string], error),
) error {

	if len(ids) == 0 {
		return nil
	}

	converted := make([]schema.ID, len(ids))
	for i := range converted {
		converted[i] = schema.ID(ids[i])
	}

	// 30 items per request.
	idsChunked := shared.ChunkSlice(converted, 30)
	for i := range idsChunked {
		if like {
			if _, err := liker(ctx, idsChunked[i]); err != nil {
				return err
			}
			continue
		}
		if _, err := unliker(ctx, idsChunked[i]); err != nil {
			return err
		}
	}

	return nil
}

type PlaylistActions struct {
	account     shared.Account
	client      *goym.Client
	myPlaylists map[shared.RemoteID]shared.RemotePlaylist
}

func (e *PlaylistActions) MyPlaylists(ctx context.Context) (map[shared.RemoteID]shared.RemotePlaylist, error) {
	if len(e.myPlaylists) > 0 {
		return e.myPlaylists, nil
	}

	playlists, err := e.client.MyPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	result := map[shared.RemoteID]shared.RemotePlaylist{}
	for i := range playlists.Result {
		// Collective playlist?
		if playlists.Result[i].Collective {
			continue
		}

		// Not user playlist?
		if playlists.Result[i].UID != e.client.UserId {
			continue
		}

		result[shared.RemoteID(playlists.Result[i].Kind.String())] = newPlaylist(e.account, playlists.Result[i], e.client)
	}

	e.myPlaylists = result
	return result, err
}

func (e PlaylistActions) Create(ctx context.Context, name string, isVisible bool, description *string) (shared.RemotePlaylist, error) {
	vis := schema.VisibilityPrivate
	if isVisible {
		vis = schema.VisibilityPublic
	}

	desc := ""
	if description != nil {
		desc = *description
	}

	pl, err := e.client.CreatePlaylist(ctx, name, desc, vis)
	if err != nil {
		return nil, err
	}

	return newPlaylist(e.account, pl.Result, e.client), err
}

func (e PlaylistActions) Delete(ctx context.Context, ids []shared.RemoteID) error {
	for _, id := range ids {
		if _, err := e.client.DeletePlaylist(ctx, schema.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

func (e PlaylistActions) Playlist(ctx context.Context, id shared.RemoteID) (shared.RemotePlaylist, error) {
	pl, err := e.client.MyPlaylist(ctx, schema.ID(id))
	notFound, err := isNotFoundOrErr(err, pl.Result.Title)
	if err != nil {
		return nil, err
	}
	if notFound {
		return nil, nil
	}
	return newPlaylist(e.account, pl.Result, e.client), err
}
