package zvuk

import (
	"context"
	"net/url"

	"github.com/oklookat/gozvuk"
	"github.com/oklookat/gozvuk/schema"
	"github.com/oklookat/synchro/shared"
)

func newTrack(track schema.Track, client *gozvuk.Client) *Track {
	return &Track{
		Entity: newEntity(track.ID.String(), track.Title),
		track:  track,
		client: client,
	}
}

type Track struct {
	*Entity
	track  schema.Track
	client *gozvuk.Client

	album shared.RemoteAlbum

	cachedArtists []shared.RemoteArtist
}

func (e Track) ISRC() *string {
	return nil
}

func (e *Track) Artists() []shared.RemoteArtist {
	if len(e.track.Artists) == 0 {
		return nil
	}
	if len(e.cachedArtists) == 0 {
		var result []shared.RemoteArtist
		for i := range e.track.Artists {
			result = append(result, newArtist(&e.track.Artists[i], e.client))
		}
		e.cachedArtists = result
	}

	return e.cachedArtists
}

func (e *Track) Album() (shared.RemoteAlbum, error) {
	if !shared.IsNil(e.album) {
		return e.album, nil
	}
	rel, err := e.client.GetReleases(context.Background(), []schema.ID{e.track.Release.ID}, 1)
	if err != nil {
		return nil, err
	}
	e.album = newAlbum(rel.Data.GetReleases[0], e.client)
	return e.album, err
}

func (e Track) LengthMs() int {
	return e.track.Duration * 1000
}

func (e *Track) Year() int {
	return e.track.Release.Date.Year()
}

func (e *Track) CoverURL() *url.URL {
	return e.track.Release.Image.SrcURL(100, 100)
}
