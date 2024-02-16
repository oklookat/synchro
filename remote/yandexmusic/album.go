package yandexmusic

import (
	"net/url"

	"github.com/oklookat/goym"
	"github.com/oklookat/goym/schema"
	"github.com/oklookat/synchro/shared"
)

func newAlbum(album schema.Album, client *goym.Client) (*Album, error) {
	if len(album.ID) == 0 {
		return nil, errEmptyID
	}

	alb := &Album{
		Entity: newEntity(album.ID.String(), album.Title),
		client: client,
		album:  album,
	}

	var cachedArtists []shared.RemoteArtist
	for i := range album.Artists {
		art, err := newArtist(album.Artists[i], client)
		if err != nil {
			return nil, err
		}
		cachedArtists = append(cachedArtists, art)
	}
	alb.cachedArtists = cachedArtists

	return alb, nil
}

type Album struct {
	*Entity
	album  schema.Album
	client *goym.Client

	cachedArtists []shared.RemoteArtist
}

func (e Album) UPC() *string {
	return nil
}

func (e Album) EAN() *string {
	return nil
}

func (e Album) Artists() []shared.RemoteArtist {
	return e.cachedArtists
}

func (e Album) TrackCount() int {
	return e.album.TrackCount
}

func (e Album) Year() int {
	return e.album.Year
}

func (e Album) CoverURL() *url.URL {
	return getCoverURL(e.album.CoverURI)
}
