package commander

import (
	"context"
	"strconv"
	"time"

	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
)

func DebugFuckup() error {
	return execTask(0, func(ctx context.Context) error {
		return repository.DebugCleanAllExceptAccounts()
	})
}

func DebugSetArtistMissing(artistID, remoteName string) error {
	aID, err := strconv.ParseUint(artistID, 10, 64)
	if err != nil {
		return err
	}
	return repository.DebugSetArtistMissing(aID, shared.RemoteName(remoteName))
}

func DebugExecInfiniteTask() error {
	return execTask(0, func(ctx context.Context) error {
		for {
			_log.Debug("DebugExecInfiniteTask() " + shared.GenerateWord())
			if ctx.Err() != nil {
				return ctx.Err()
			}
			time.Sleep(100 * time.Millisecond)
		}
	})
}
