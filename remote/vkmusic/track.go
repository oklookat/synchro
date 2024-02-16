package vkmusic

import (
	"context"
	"net/url"

	"github.com/oklookat/govkm"
	"github.com/oklookat/govkm/schema"
	"github.com/oklookat/synchro/shared"
)

func newTrack(track schema.Track, client *govkm.Client) (*Track, error) {
	full, err := client.Album(context.Background(), track.Album.APIID)
	if err != nil {
		return nil, err
	}
	if full.Data.Album == nil {
		return nil, errNilAlbum
	}

	return &Track{
		Entity: newEntity(track.APIID.String(), track.Name),
		track:  track,
		client: client,

		cachedAlbum: *full.Data.Album,
	}, err
}

type Track struct {
	*Entity
	track  schema.Track
	client *govkm.Client

	cachedAlbum schema.Album
}

func (e Track) ISRC() *string {
	return nil
}

func (e *Track) Artists() []shared.RemoteArtist {
	var artists []schema.SimpleArtist
	artists = append(artists, e.track.Artists...)

	var result []shared.RemoteArtist
	for i := range artists {
		result = append(result, newArtist(artists[i], e.client))
	}

	return result
}

func (e *Track) Album() (shared.RemoteAlbum, error) {
	return newAlbum(e.cachedAlbum, e.client), nil
}

func (e Track) LengthMs() int {
	return e.track.Duration * 1000
}

func (e Track) Year() int {
	return e.cachedAlbum.Year
}

func (e Track) CoverURL() *url.URL {
	return e.track.Cover.GetUrl()
}
