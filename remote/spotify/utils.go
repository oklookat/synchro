package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/oklookat/synchro/shared"
	"github.com/zmb3/spotify/v2"
)

func authorizedToAuth(au *authorized) (string, error) {
	byted, err := json.Marshal(au)
	if err != nil {
		return "", err
	}
	return string(byted), err
}

func authToAuthorized(au string) (*authorized, error) {
	auth := &authorized{}
	err := json.Unmarshal([]byte(au), auth)
	return auth, err
}

func newEntity(id, name string) *Entity {
	return &Entity{
		id:   id,
		name: name,
	}
}

type Entity struct {
	id   string
	name string
}

func (e Entity) RemoteName() shared.RemoteName {
	return _repo.Name()
}

func (e Entity) ID() shared.RemoteID {
	return shared.RemoteID(e.id)
}

func (e Entity) Name() string {
	return e.name
}

func getCoverURL(images []spotify.Image) *url.URL {
	if len(images) == 0 {
		return nil
	}

	imageURL := ""
	for i := range images {
		if images[i].Height <= 100 {
			imageURL = images[i].URL
			break
		}
	}

	if len(imageURL) == 0 {
		return nil
	}

	url, err := url.Parse(imageURL)
	if err != nil {
		return nil
	}
	return url
}

// Search. If 404, makes 20 attempts to get a result.
//
// Spotify API can accidentally throw up a 404 when searching.
//
// I haven't figured out why yet. Perhaps the Market is related to this.
// Maybe from some IPs (for example, Russian IPs) the search behaves strangely. Idk.
func pleaseSearch(ctx context.Context, cl *spotify.Client, query string, t spotify.SearchType, opts ...spotify.RequestOption) (result *spotify.SearchResult, err error) {
	for i := 0; i < 20; i++ {
		result, err = cl.Search(ctx, query, t, opts...)
		if err == nil {
			return result, nil
		}
		if isNotFound(err) {
			time.Sleep(350 * time.Millisecond)
		}
	}
	if isNotFound(err) {
		return result, nil
	}
	return result, err
}

func isNotFound(err error) bool {
	spotErr := &spotify.Error{}
	return errors.As(err, spotErr) && spotErr.Status == 404
}
