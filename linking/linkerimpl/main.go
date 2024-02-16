package linkerimpl

import (
	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/shared"
)

var (
	_log     *logger.Logger
	_remotes map[shared.RemoteName]shared.Remote
)

const (
	_packageName = "linkerimpl"
)

func Boot(remotes map[shared.RemoteName]shared.Remote) {
	_remotes = remotes
	_log = logger.WithPackageName(_packageName)
}

func NewRemoteEntity(from shared.RemoteEntity) linker.RemoteEntity {
	return from.(linker.RemoteEntity)
}

// Get remotes that ready to working with linker
// (accounts can send api requests, etc).
func checkRemotes() map[shared.RemoteName]shared.Remote {
	ready := map[shared.RemoteName]shared.Remote{}
	for name := range _remotes {
		if !_remotes[name].Repository().Enabled() {
			continue
		}
		_, err := _remotes[name].Actions()
		if err != nil {
			continue
		}
		ready[name] = _remotes[name]
	}
	return ready
}
