package commander

import (
	"context"

	"github.com/oklookat/synchro/syncing/lisybridge"
)

// Start syncing.
func Sync() error {
	return execTask(0, func(ctx context.Context) error {
		return lisybridge.Sync(ctx)
	})
}
