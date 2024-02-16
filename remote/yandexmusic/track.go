package yandexmusic

import (
	"net/url"

	"github.com/oklookat/goym"
	"github.com/oklookat/goym/schema"
	"github.com/oklookat/synchro/shared"
)

func newTrack(track schema.Track, client *goym.Client) (*Track, error) {
	if len(track.ID) == 0 {
		return nil, errEmptyID
	}

	tr := &Track{
		Entity: newEntity(track.ID.String(), track.Title),
		track:  track,
		client: client,
	}

	var cachedArtists []shared.RemoteArtist
	for i := range track.Artists {
		art, err := newArtist(track.Artists[i], client)
		if err != nil {
			return nil, err
		}
		cachedArtists = append(cachedArtists, art)
	}
	tr.cachedArtists = cachedArtists

	return tr, nil
}

type Track struct {
	*Entity
	track  schema.Track
	client *goym.Client

	cachedArtists []shared.RemoteArtist
	cachedAlbum   shared.RemoteAlbum
}

func (e Track) ISRC() *string {
	return nil
}

func (e *Track) Artists() []shared.RemoteArtist {
	return e.cachedArtists
}

func (e *Track) Album() (shared.RemoteAlbum, error) {
	if !shared.IsNil(e.cachedAlbum) {
		return e.cachedAlbum, nil
	}

	alb, err := newAlbum(e.track.Albums[0], e.client)
	e.cachedAlbum = alb
	return e.cachedAlbum, err
}

func (e Track) LengthMs() int {
	return e.track.DurationMs
}

func (e Track) Year() int {
	if len(e.track.Albums) == 0 {
		return 0
	}
	return e.track.Albums[0].Year
}

func (e Track) CoverURL() *url.URL {
	return getCoverURL(e.track.CoverUri)
}
