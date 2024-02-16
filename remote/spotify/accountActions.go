package spotify

import (
	"context"

	"github.com/oklookat/synchro/shared"
	"github.com/zmb3/spotify/v2"
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
	client  *spotify.Client
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
	client *spotify.Client
}

func (e LikedAlbumsActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	albums := map[shared.RemoteID]shared.RemoteEntity{}
	offset := 0

	for {
		albumsd, err := e.client.CurrentUsersAlbums(ctx, spotify.Limit(45), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		for i := range albumsd.Albums {
			albums[shared.RemoteID(albumsd.Albums[i].ID)] = newAlbum(&albumsd.Albums[i].FullAlbum, e.client)
		}

		if len(albums) >= albumsd.Total {
			break
		}

		offset += albumsd.Limit
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
	return likeUnlike(ctx, ids, like, e.client.AddAlbumsToLibrary, e.client.RemoveAlbumsFromLibrary)
}

type LikedArtistsActions struct {
	client *spotify.Client
}

func (e LikedArtistsActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	artists := map[shared.RemoteID]shared.RemoteEntity{}
	var offset string

	for {
		var options []spotify.RequestOption
		options = append(options, spotify.Limit(45))
		if len(offset) > 0 {
			options = append(options, spotify.After(offset))
		}

		followed, err := e.client.CurrentUsersFollowedArtists(ctx, options...)
		if err != nil {
			return nil, err
		}

		for i := range followed.Artists {
			artists[shared.RemoteID(followed.Artists[i].ID)] = newArtist(followed.Artists[i].SimpleArtist, e.client)
		}

		if len(followed.Cursor.After) == 0 {
			break
		}

		offset = followed.Cursor.After
	}

	return artists, nil
}

func (e LikedArtistsActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedArtistsActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedArtistsActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	return likeUnlike(ctx, ids, like, e.client.FollowArtist, e.client.UnfollowArtist)
}

type LikedTracksActions struct {
	client *spotify.Client
}

func (e LikedTracksActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	tracks := map[shared.RemoteID]shared.RemoteEntity{}
	offset := 0

	for {
		currentUser, err := e.client.CurrentUsersTracks(ctx, spotify.Limit(45), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		for i := range currentUser.Tracks {
			tracks[shared.RemoteID(currentUser.Tracks[i].ID)] = newTrack(currentUser.Tracks[i].FullTrack, e.client)
		}

		if len(tracks) >= currentUser.Total || len(currentUser.Next) == 0 {
			break
		}

		offset += currentUser.Limit
	}

	return tracks, nil
}

func (e LikedTracksActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, true)
}

func (e LikedTracksActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.likeUnlike(ctx, ids, false)
}

func (e LikedTracksActions) likeUnlike(ctx context.Context, ids []shared.RemoteID, like bool) error {
	return likeUnlike(ctx, ids, like, e.client.AddTracksToLibrary, e.client.RemoveTracksFromLibrary)
}

func likeUnlike(
	ctx context.Context,
	ids []shared.RemoteID,
	like bool,
	liker func(ctx context.Context, ids ...spotify.ID) error,
	unliker func(ctx context.Context, ids ...spotify.ID) error,
) error {

	converted := make([]spotify.ID, len(ids))
	for i := range converted {
		converted[i] = spotify.ID(ids[i].String())
	}

	// 25 items per request.
	idsChunked := shared.ChunkSlice(converted, 25)
	for i := range idsChunked {
		if like {
			if err := liker(ctx, idsChunked[i]...); err != nil {
				return err
			}
			continue
		}
		if err := unliker(ctx, idsChunked[i]...); err != nil {
			return err
		}
	}

	return nil
}

type PlaylistActions struct {
	account     shared.Account
	client      *spotify.Client
	currentUser *spotify.PrivateUser
}

func (e *PlaylistActions) MyPlaylists(ctx context.Context) (map[shared.RemoteID]shared.RemotePlaylist, error) {
	if err := e.cache(ctx); err != nil {
		return nil, err
	}

	result := map[shared.RemoteID]shared.RemotePlaylist{}

	const limit = 45
	offset := 0

	for {
		page, err := e.client.GetPlaylistsForUser(
			ctx,
			e.currentUser.ID,
			spotify.Limit(limit),
			spotify.Offset(offset))
		if err != nil {
			return nil, err
		}
		for _, item := range page.Playlists {
			if item.Collaborative {
				continue
			}
			if item.Owner.ID != e.currentUser.ID {
				continue
			}
			result[shared.RemoteID(item.ID)] = newPlaylist(e.account, item, e.client)
		}
		if len(page.Next) == 0 {
			break
		}
		offset += limit
	}

	return result, nil
}

func (e PlaylistActions) Create(ctx context.Context, name string, isVisible bool, description *string) (shared.RemotePlaylist, error) {
	if err := e.cache(ctx); err != nil {
		return nil, err
	}

	desc := ""
	if description != nil {
		desc = *description
	}

	pl, err := e.client.CreatePlaylistForUser(ctx, e.currentUser.ID, name, desc, isVisible, false)
	if err != nil {
		return nil, err
	}

	return newPlaylist(e.account, pl.SimplePlaylist, e.client), err
}

func (e PlaylistActions) Delete(ctx context.Context, entities []shared.RemoteID) error {
	for _, id := range entities {
		if err := e.client.UnfollowPlaylist(ctx, spotify.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

func (e PlaylistActions) Playlist(ctx context.Context, id shared.RemoteID) (shared.RemotePlaylist, error) {
	pl, err := e.client.GetPlaylist(ctx, spotify.ID(id))
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return newPlaylist(e.account, pl.SimplePlaylist, e.client), err
}

func (e *PlaylistActions) cache(ctx context.Context) error {
	if e.currentUser != nil {
		return nil
	}
	usr, err := e.client.CurrentUser(ctx)
	if err != nil {
		return err
	}
	e.currentUser = usr
	return err
}
