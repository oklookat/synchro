package linkerimpl

import (
	"context"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/streaming"
)

var (
	_remotes map[streaming.ServiceName]streaming.Service
)

const (
	_packageName = "linkerimpl"
)

func Boot(services map[streaming.ServiceName]streaming.Service) {
	_remotes = services
}

func NewRemoteEntity(from streaming.ServiceEntity) linker.RemoteEntity {
	return from.(linker.RemoteEntity)
}

// Get services that ready to working with linker
// (accounts can send api requests, etc).
func checkRemotes() map[streaming.ServiceName]streaming.Service {
	ready := map[streaming.ServiceName]streaming.Service{}
	for name, rem := range _remotes {
		accs, err := rem.Database().Accounts(context.Background())
		if err != nil || len(accs) == 0 {
			continue
		}
		if _, err = rem.Actions(); err != nil {
			continue
		}
		ready[name] = _remotes[name]
	}
	return ready
}
