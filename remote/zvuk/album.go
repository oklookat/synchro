package zvuk

import (
	"net/url"

	"github.com/oklookat/gozvuk"
	"github.com/oklookat/gozvuk/schema"
	"github.com/oklookat/synchro/shared"
)

func newAlbum(album schema.Release, client *gozvuk.Client) *Album {
	return &Album{
		Entity: newEntity(album.ID.String(), album.Title),
		client: client,
		album:  album,
	}
}

type Album struct {
	*Entity

	client *gozvuk.Client
	album  schema.Release
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
		result = append(result, newArtist(&e.album.Artists[i], e.client))
	}
	return result
}

func (e Album) TrackCount() int {
	return len(e.album.Tracks)
}

func (e Album) Year() int {
	return e.album.Date.Year()
}

func (e Album) CoverURL() *url.URL {
	return e.album.Image.SrcURL(100, 100)
}
