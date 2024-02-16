package zvuk

import (
	"context"

	"github.com/oklookat/gozvuk"
	"github.com/oklookat/gozvuk/schema"
	"github.com/oklookat/synchro/shared"
)

func newActions(client *gozvuk.Client) *Actions {
	return &Actions{
		client: client,
	}
}

type Actions struct {
	client *gozvuk.Client
}

func (e Actions) Album(ctx context.Context, id shared.RemoteID) (shared.RemoteAlbum, error) {
	if len(id) == 0 {
		return nil, nil
	}
	album, err := e.client.GetReleases(ctx, []schema.ID{schema.ID(id)}, 1)
	if err != nil {
		return nil, err
	}
	if len(album.Data.GetReleases) == 0 || len(album.Data.GetReleases[0].ID) == 0 {
		return nil, nil
	}
	return newAlbum(album.Data.GetReleases[0], e.client), nil
}

func (e Actions) Artist(ctx context.Context, id shared.RemoteID) (shared.RemoteArtist, error) {
	if len(id) == 0 {
		return nil, nil
	}
	resp, err := e.client.GetArtists(ctx, []schema.ID{schema.ID(id)}, false, 1, 0, false, 1, 0, false, 1, false)
	if err != nil {
		return nil, err
	}
	if len(resp.Data.GetArtists) == 0 || len(resp.Data.GetArtists[0].ID) == 0 {
		return nil, nil
	}
	return newArtist(&resp.Data.GetArtists[0].SimpleArtist, e.client), err
}

func (e Actions) Track(ctx context.Context, id shared.RemoteID) (shared.RemoteTrack, error) {
	if len(id) == 0 {
		return nil, nil
	}
	trackd, err := e.client.GetFullTrack(ctx, []schema.ID{schema.ID(id)})
	if err != nil {
		return nil, err
	}

	if len(trackd.Data.GetTracks) == 0 || len(trackd.Data.GetTracks[0].ID) == 0 {
		return nil, nil
	}

	return newTrack(trackd.Data.GetTracks[0], e.client), err
}

func (e Actions) SearchAlbums(ctx context.Context, what shared.RemoteAlbum) ([10]shared.RemoteAlbum, error) {
	action := &AlbumsSearchAction{e.client}
	return action.Search(ctx, what)
}

func (e Actions) SearchArtists(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	action := &ArtistsSearchAction{e.client}
	return action.Search(ctx, what)
}

func (e Actions) SearchTracks(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	action := &TracksSearchAction{e.client}
	return action.Search(ctx, what)
}

type AlbumsSearchAction struct {
	client *gozvuk.Client
}

func (e AlbumsSearchAction) Search(ctx context.Context, what shared.RemoteAlbum) ([10]shared.RemoteAlbum, error) {
	artistName := what.Artists()[0].Name()
	albumName := what.Name()

	query := artistName + " " + shared.SearchablePart(albumName)
	result, err := e.search(ctx, query)
	if err != nil || shared.IsNil(result[0]) {
		// Error or not found.
		return result, err
	}

	// Try search again (normalize + suggest).
	query = shared.SearchableNormalized(artistName, albumName)
	return e.search(ctx, query)
}

func (e AlbumsSearchAction) search(ctx context.Context, query string) ([10]shared.RemoteAlbum, error) {
	var (
		result [10]shared.RemoteAlbum
	)

	search, err := e.client.Search(ctx, schema.SearchArguments{
		Query:    query,
		Limit:    10,
		Releases: true,
	})

	if err != nil {
		return result, err
	}

	releases := search.Data.Search.Releases

	if releases == nil ||
		len(releases.Items) == 0 ||
		len(releases.Items[0].ID) == 0 {
		return result, nil
	}

	var releasesIds []schema.ID
	for _, item := range releases.Items {
		releasesIds = append(releasesIds, item.ID)
	}

	fullRel, err := e.client.GetReleases(ctx, releasesIds, 1)
	if err != nil {
		return result, err
	}

	for i := range result {
		if i == len(fullRel.Data.GetReleases) {
			break
		}
		result[i] = newAlbum(fullRel.Data.GetReleases[i], e.client)
	}

	return result, err
}

type ArtistsSearchAction struct {
	client *gozvuk.Client
}

func (e ArtistsSearchAction) Search(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	var result [10]shared.RemoteArtist

	search, err := e.client.Search(ctx, schema.SearchArguments{
		Query:   what.Name(),
		Artists: true,
		Limit:   10,
	})
	if err != nil {
		return result, err
	}

	artists := search.Data.Search.Artists

	if artists == nil || len(artists.Items) == 0 || len(artists.Items[0].ID) == 0 {
		return result, nil
	}

	for i := range result {
		if i == len(artists.Items) {
			break
		}
		result[i] = newArtist(&artists.Items[i], e.client)
	}

	return result, err
}

type TracksSearchAction struct {
	client *gozvuk.Client
}

func (e TracksSearchAction) Search(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	var result [10]shared.RemoteTrack

	search, err := e.client.Search(ctx, schema.SearchArguments{
		Query:  what.Artists()[0].Name() + " " + shared.SearchablePart2(what.Name()),
		Tracks: true,
		Limit:  10,
	})
	if err != nil {
		return result, err
	}

	tracks := search.Data.Search.Tracks

	if tracks == nil || len(tracks.Items) == 0 || len(tracks.Items[0].ID) == 0 {
		return result, nil
	}

	var tracksIds []schema.ID
	for _, item := range tracks.Items {
		tracksIds = append(tracksIds, item.ID)
	}

	fullTracks, err := e.client.GetFullTrack(ctx, tracksIds)
	if err != nil {
		return result, err
	}

	for i := range result {
		if i == len(fullTracks.Data.GetTracks) {
			break
		}
		result[i] = newTrack(fullTracks.Data.GetTracks[i], e.client)
	}

	return result, err
}
