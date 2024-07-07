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
