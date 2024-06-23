package spotify

import (
	"net/url"

	"github.com/oklookat/synchro/shared"
	"github.com/zmb3/spotify/v2"
)

func newAlbum(album *spotify.FullAlbum, client *spotify.Client) *Album {
	return &Album{
		Entity: newEntity(album.ID.String(), album.Name),
		client: client,
		album:  album,
	}
}

type Album struct {
	*Entity
	album  *spotify.FullAlbum
	client *spotify.Client
}

func (e Album) UPC() *string {
	if len(e.album.ExternalIDs) == 0 {
		return nil
	}
	upc, ok := e.album.ExternalIDs["upc"]
	if !ok {
		return nil
	}
	return &upc
}

func (e Album) EAN() *string {
	if len(e.album.ExternalIDs) == 0 {
		return nil
	}
	ean, ok := e.album.ExternalIDs["ean"]
	if !ok {
		return nil
	}
	return &ean
}

func (e Album) Artists() []shared.RemoteArtist {
	if len(e.album.Artists) == 0 {
		return nil
	}

	var result []shared.RemoteArtist
	for i := range e.album.Artists {
		result = append(result, newArtist(e.album.Artists[i], e.client))
	}

	return result
}

func (e Album) TrackCount() int {
	return int(e.album.Tracks.Total)
}

func (e Album) Year() int {
	return e.album.ReleaseDateTime().Year()
}

func (e Album) CoverURL() *url.URL {
	return getCoverURL(e.album.Images)
}
