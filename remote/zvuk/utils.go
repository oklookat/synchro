package zvuk

import (
	"time"

	"github.com/oklookat/synchro/shared"
	"golang.org/x/oauth2"
)

func newToken(accessToken string) *oauth2.Token {
	return &oauth2.Token{
		TokenType:    "Bearer",
		AccessToken:  accessToken,
		RefreshToken: "",
		Expiry:       time.Now(),
	}
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
