package vkmusic

import (
	"net/url"

	"github.com/oklookat/govkm"
	"github.com/oklookat/govkm/schema"
	"github.com/oklookat/synchro/shared"
)

func newAlbum(album schema.Album, client *govkm.Client) *Album {
	return &Album{
		Entity: newEntity(album.APIID.String(), album.Name),
		client: client,
		album:  album,
	}
}

type Album struct {
	*Entity
	album  schema.Album
	client *govkm.Client
}

func (e Album) UPC() *string {
	return nil
}

func (e Album) EAN() *string {
	return nil
}

func (e Album) Artists() []shared.RemoteArtist {
	if len(e.album.Artists) == 0 {
		return nil
	}

	var result []shared.RemoteArtist
	for i := range e.album.Artists {
		result = append(result, newArtist(e.album.Artists[i].SimpleArtist, e.client))
	}

	return result
}

func (e Album) TrackCount() int {
	return e.album.Counts.Track
}

func (e Album) Year() int {
	return e.album.ReleaseDateTimestamp.Time().Year()
}

func (e Album) CoverURL() *url.URL {
	return e.album.Cover.GetUrl()
}
