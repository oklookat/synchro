package linkerimpl

import (
	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
)

var (
	_remotes map[shared.RemoteName]shared.Remote
)

func Boot(remotes map[shared.RemoteName]shared.Remote) {
	_remotes = remotes
}

func NewRemoteEntity(from shared.RemoteEntity) linker.RemoteEntity {
	return from.(linker.RemoteEntity)
}

// Get remotes that ready to working with linker
// (accounts can send api requests, etc).
func checkRemotes() map[shared.RemoteName]shared.Remote {
	ready := map[shared.RemoteName]shared.Remote{}
	for name := range _remotes {
		acts, err := _remotes[name].Actions()
		if err != nil {
			continue
		}
		if shared.IsNil(acts) {
			// Actions not available / no accounts, remote disabled, etc.
			continue
		}
		ready[name] = _remotes[name]
	}
	return ready
}
