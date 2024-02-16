package deezer

import (
	"context"
	"net/url"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/schema"
	"github.com/oklookat/synchro/shared"
)

func newAlbum(ctx context.Context, cl *deezus.Client, id schema.ID) (*Album, error) {
	resp, err := cl.Album(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &Album{
		Entity: newEntity(resp.ID.String(), resp.Title),
		client: cl,
		album:  resp.Album,
	}, err
}

type Album struct {
	*Entity
	album  schema.Album
	client *deezus.Client
}

func (e Album) UPC() *string {
	if len(e.album.Upc) == 0 {
		return nil
	}
	return &e.album.Upc
}

func (e Album) EAN() *string {
	return nil
}

func (e Album) Artists() []shared.RemoteArtist {
	var result []shared.RemoteArtist
	for i := range e.album.Contributors {
		result = append(result,
			newArtist(
				e.client,
				e.album.Contributors[i].SimpleArtist))
	}

	return result
}

func (e Album) TrackCount() int {
	return e.album.NbTracks
}

func (e Album) Year() int {
	if e.album.ReleaseDate == nil {
		return -1
	}
	return e.album.ReleaseDate.Time().Year()
}

func (e Album) CoverURL() *url.URL {
	return getCoverURL(e.album.Cover)
}
