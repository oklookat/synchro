package yandexmusic

import (
	"errors"
	"net/url"
	"strings"

	"github.com/oklookat/goym/schema"
	"github.com/oklookat/synchro/shared"
)

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

func getCoverURL(urld string) *url.URL {
	if len(urld) == 0 {
		return nil
	}
	converted := strings.TrimSuffix(urld, "%%") + "m100x100"
	url, err := url.Parse("https://" + converted)
	if err != nil {
		return nil
	}
	return url
}

func isNotFound(err error) bool {
	var respErr schema.Error
	if errors.As(err, &respErr) {
		return respErr.IsNotFound() || respErr.IsValidate()
	}
	return false
}

func isNotFoundOrErr(err error, strCheck string) (bool, error) {
	if err == nil {
		if len(strCheck) == 0 {
			return true, nil
		}
	}
	if isNotFound(err) {
		return true, nil
	}
	return false, err
}

func isUgcTrack(tr schema.Track) bool {
	return tr.TrackSource == schema.TrackSourceUgc || (tr.Filename != nil && len(*tr.Filename) > 0)
}
