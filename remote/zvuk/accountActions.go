package zvuk

import (
	"context"
	"errors"
	"strconv"

	"github.com/oklookat/gozvuk"
	"github.com/oklookat/gozvuk/schema"
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
	client  *gozvuk.Client
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
	client *gozvuk.Client
}

func (e LikedAlbumsActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	col, err := e.client.UserCollection(ctx)
	if err != nil {
		return nil, err
	}

	releases := col.Data.Collection.Releases
	var albumIds []schema.ID
	for _, ci := range releases {
		if ci.ID == nil {
			continue
		}
		albumIds = append(albumIds, *ci.ID)
	}

	result := map[shared.RemoteID]shared.RemoteEntity{}
	// 30 items per request.
	idsChunked := shared.ChunkSlice(albumIds, 30)
	for i := range idsChunked {
		albumd, err := e.client.GetReleases(ctx, idsChunked[i], 1)
		if err != nil {
			return nil, err
		}
		for x := range albumd.Data.GetReleases {
			id := albumd.Data.GetReleases[x].ID
			if len(id) == 0 {
				continue
			}
			result[shared.RemoteID(id)] = newAlbum(albumd.Data.GetReleases[x], e.client)
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
	return likeUnlike(ctx, e.client, ids, schema.CollectionItemTypeRelease, like)
}

type LikedArtistsActions struct {
	client *gozvuk.Client
}

func (e LikedArtistsActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	collResp, err := e.client.UserCollection(ctx)
	if err != nil {
		return nil, err
	}

	var ids []schema.ID
	for _, item := range collResp.Data.Collection.Artists {
		if item.ID == nil || len(*item.ID) == 0 {
			continue
		}
		ids = append(ids, *item.ID)
	}

	var artists []schema.Artist
	idsChunks := shared.ChunkSlice(ids, 20)
	for _, chunk := range idsChunks {
		artResp, err := e.client.GetArtists(ctx, chunk, false, 1, 0, false, 1, 0, false, 1, false)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artResp.Data.GetArtists...)
	}

	result := map[shared.RemoteID]shared.RemoteEntity{}
	for i := range artists {
		if len(artists[i].ID) == 0 {
			continue
		}
		result[shared.RemoteID(artists[i].ID.String())] = newArtist(&artists[i].SimpleArtist, e.client)
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
	return likeUnlike(ctx, e.client, ids, schema.CollectionItemTypeArtist, like)
}

type LikedTracksActions struct {
	client *gozvuk.Client
}

func (e LikedTracksActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	resp, err := e.client.UserTracks(ctx, schema.OrderByDateAdded, schema.OrderDirectionAsc)
	if err != nil {
		return nil, err
	}
	if len(resp.Data.Collection.Tracks) == 0 ||
		len(resp.Data.Collection.Tracks[0].ID) == 0 {
		return nil, err
	}

	var trackIds []schema.ID
	for _, v := range resp.Data.Collection.Tracks {
		trackIds = append(trackIds, v.ID)
	}

	result := map[shared.RemoteID]shared.RemoteEntity{}

	// 30 items per request.
	idsChunked := shared.ChunkSlice(trackIds, 10)
	for i := range idsChunked {
		trackd, err := e.client.GetFullTrack(ctx, idsChunked[i])
		if err != nil {
			return nil, err
		}
		for x := range trackd.Data.GetTracks {
			if len(trackd.Data.GetTracks[x].ID) == 0 {
				// Track not found? Zvuk sometimes gives not exists tracks. Idk why.
				continue
			}
			result[shared.RemoteID(trackd.Data.GetTracks[x].ID.String())] = newTrack(trackd.Data.GetTracks[x], e.client)
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
	return likeUnlike(ctx, e.client, ids, schema.CollectionItemTypeTrack, like)
}

func likeUnlike(ctx context.Context, cl *gozvuk.Client, ids []shared.RemoteID, itype schema.CollectionItemType, like bool) error {
	var converted []schema.ID
	for _, id := range ids {
		if len(id) == 0 {
			continue
		}
		converted = append(converted, schema.ID(id))
	}

	for _, id := range converted {
		if like {
			_, err := cl.AddItemToCollection(ctx, id, itype)
			if err != nil {
				return err
			}
			continue
		}
		_, err := cl.RemoveItemFromCollection(ctx, id, itype)
		if err != nil {
			return err
		}
	}

	return nil
}

type PlaylistActions struct {
	profileID   string
	account     shared.Account
	client      *gozvuk.Client
	myPlaylists map[shared.RemoteID]shared.RemotePlaylist
}

func (e *PlaylistActions) MyPlaylists(ctx context.Context) (map[shared.RemoteID]shared.RemotePlaylist, error) {
	if len(e.myPlaylists) > 0 {
		return e.myPlaylists, nil
	}
	resp, err := e.client.UserPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	prof, err := e.client.Profile()
	if err != nil {
		return nil, err
	}
	if prof.Result.ID == nil {
		return nil, errors.New("nil profile id")
	}
	e.profileID = strconv.Itoa(*prof.Result.ID)

	var myPlaylistsIds []schema.ID
	for _, item := range resp.Data.Collection.Playlists {
		if item.UserID == nil || item.ID == nil {
			continue
		}
		myPlaylistsIds = append(myPlaylistsIds, *item.ID)
	}

	result := map[shared.RemoteID]shared.RemotePlaylist{}
	chunksIds := shared.ChunkSlice(myPlaylistsIds, 10)
	for _, chunk := range chunksIds {
		playlists, err := e.client.GetPlaylists(ctx, chunk)
		if err != nil {
			return nil, err
		}
		playlistsSlice := playlists.Data.GetPlaylists
		for i := range playlistsSlice {
			id := playlistsSlice[i].ID
			if len(id) == 0 {
				continue
			}
			result[shared.RemoteID(id)] = newPlaylist(e.account, playlistsSlice[i], e.client)
		}
	}

	e.myPlaylists = result
	return result, err
}

func (e PlaylistActions) Create(ctx context.Context, name string, isVisible bool, description *string) (shared.RemotePlaylist, error) {
	resp, err := e.client.CreatePlaylist(ctx, []schema.PlaylistItem{
		// We can't create playlists without tracks.
		schema.NewPlaylistItem(schema.PlaylistItemTypeTrack, _tempPlaylistTrackId),
	}, name)
	if err != nil {
		return nil, err
	}
	id := resp.Data.Playlist.Create

	_, err = e.client.SetPlaylistToPublic(ctx, id, isVisible)
	if err != nil {
		return nil, err
	}

	plResp, err := e.client.GetPlaylists(ctx, []schema.ID{id})
	if err != nil {
		return nil, err
	}

	return newPlaylist(e.account, plResp.Data.GetPlaylists[0], e.client), err
}

func (e PlaylistActions) Delete(ctx context.Context, ids []shared.RemoteID) error {
	for _, id := range ids {
		if len(id) == 0 {
			continue
		}
		if _, err := e.client.DeletePlaylist(ctx, schema.ID(id)); err != nil {
			return err
		}
	}
	return nil
}

func (e PlaylistActions) Playlist(ctx context.Context, id shared.RemoteID) (shared.RemotePlaylist, error) {
	pl, err := e.client.GetPlaylists(ctx, []schema.ID{schema.ID(id)})
	if err != nil {
		return nil, err
	}
	if len(pl.Data.GetPlaylists) == 0 || len(pl.Data.GetPlaylists[0].ID) == 0 {
		// Not found.
		return nil, nil
	}
	return newPlaylist(e.account, pl.Data.GetPlaylists[0], e.client), err
}
