package linker

import "github.com/oklookat/synchro/streaming"

func Boot() {

}

var ()

const (
	_packageName = "linker"
)

type (
	// Repository of service-independent entity.
	//
	// Example: artist repository.
	Repository interface {
		CreateEntity() (streaming.DatabaseEntityID, error)

		// 1. Delete entities that not linked with any service.
		//
		// 2. Delete entities that linked, but have NULL RemoteID on all services.
		DeleteNotLinked() error

		// Delete all entities.
		DeleteAll() error
	}
)
