package deezer

import (
	"context"
	"net/url"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/schema"
	"github.com/oklookat/synchro/shared"
)

func newTrack(ctx context.Context, cl *deezus.Client, id schema.ID) (*Track, error) {
	resp, err := cl.Track(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &Track{
		Entity: newEntity(resp.ID.String(), resp.Title),
		track:  resp.Track,
		client: cl,
	}, err
}

type Track struct {
	*Entity
	track  schema.Track
	client *deezus.Client

	cachedAlbum   *Album
	cachedArtists []shared.RemoteArtist
}

func (e Track) ISRC() *string {
	return &e.track.Isrc
}

func (e Track) Artists() []shared.RemoteArtist {
	if len(e.cachedArtists) > 0 {
		return e.cachedArtists
	}
	for i := range e.track.Contributors {
		e.cachedArtists = append(e.cachedArtists, newArtist(e.client, e.track.Contributors[i].SimpleArtist))
	}
	return e.cachedArtists
}

func (e *Track) Album() (shared.RemoteAlbum, error) {
	if e.cachedAlbum != nil {
		return e.cachedAlbum, nil
	}
	al, err := newAlbum(context.Background(), e.client, e.track.Album.ID)
	if err != nil {
		return nil, err
	}
	e.cachedAlbum = al
	return al, err
}

func (e Track) LengthMs() int {
	return e.track.Duration * 1000
}

func (e Track) Year() int {
	return e.track.ReleaseDate.Time().Year()
}

func (e Track) CoverURL() *url.URL {
	return getCoverURL(e.track.Album.Cover)
}
