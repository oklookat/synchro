package spotify

import (
	"context"
	"net/url"

	"github.com/oklookat/synchro/shared"
	"github.com/zmb3/spotify/v2"
)

func newTrack(track spotify.FullTrack, client *spotify.Client) *Track {
	return &Track{
		Entity: newEntity(track.ID.String(), track.Name),
		track:  track,
		client: client,
	}
}

type Track struct {
	*Entity
	track  spotify.FullTrack
	client *spotify.Client

	artists     []shared.RemoteArtist
	cachedAlbum *spotify.FullAlbum
}

func (e Track) ISRC() *string {
	if len(e.track.ExternalIDs) == 0 {
		return nil
	}
	isrc, ok := e.track.ExternalIDs["isrc"]
	if !ok {
		return nil
	}
	return &isrc
}

func (e *Track) Artists() []shared.RemoteArtist {
	if len(e.track.Artists) == 0 {
		return nil
	}
	if len(e.artists) == 0 {
		var result []shared.RemoteArtist
		for i := range e.track.Artists {
			result = append(result, newArtist(e.track.Artists[i], e.client))
		}
		e.artists = result
	}
	return e.artists
}

func (e *Track) Album() (shared.RemoteAlbum, error) {
	if e.cachedAlbum != nil {
		return newAlbum(e.cachedAlbum, e.client), nil
	}
	full, err := e.client.GetAlbum(context.Background(), e.track.Album.ID)
	e.cachedAlbum = full
	return newAlbum(e.cachedAlbum, e.client), err
}

func (e Track) LengthMs() int {
	return e.track.Duration
}

func (e Track) Year() int {
	return e.track.Album.ReleaseDateTime().Year()
}

func (e Track) CoverURL() *url.URL {
	return getCoverURL(e.track.Album.Images)
}
