package linker

import (
	"github.com/oklookat/synchro/shared"
)

type (
	// Repository of remote-independent entity.
	//
	// Example: artist repository.
	Repository interface {
		CreateEntity() (shared.EntityID, error)

		// 1. Delete entities that not linked with any remote.
		//
		// 2. Delete entities that linked, but have NULL RemoteID on all remotes.
		DeleteNotLinked() error

		// Delete all entities.
		DeleteAll() error
	}

	// Can import/export links.
	RepositoryShareable interface {
		Repository
	}
)
