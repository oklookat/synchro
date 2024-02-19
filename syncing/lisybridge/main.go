package lisybridge

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/shared"
)

var (
	_isSyncing bool

	_log     *logger.Logger
	_remotes map[shared.RemoteName]shared.Remote
)

type (
	Account interface {
		RemoteName() shared.RemoteName
		Alias() string
		ID() string
		LastSynchronization() time.Time
		SetLastSynchronization(time.Time) error
	}
)

func Boot(remotes map[shared.RemoteName]shared.Remote) {
	_log = logger.WithPackageName("lisybridge")
	_remotes = remotes
}

func Sync(ctx context.Context) error {
	if _isSyncing {
		return errors.New("another sync in progress")
	}

	_isSyncing = true
	defer func() {
		_isSyncing = false
		onSyncEnd(ctx)
		_log.Info("Done.")
	}()

	_log.Info("Executing start hook...")
	if err := onSyncStart(ctx); err != nil {
		return err
	}

	_log.Info("Collecting syncable accounts...")
	accs, err := getAccountsForSync(ctx)
	if err != nil {
		return err
	}

	if len(accs.LikedAlbums) > 0 {
		_log.
			AddField("accountsCount", strconv.Itoa(len(accs.LikedAlbums))).
			Info("Syncing liked albums...")
		if err := syncLikedAlbums(ctx, accs.LikedAlbums); err != nil {
			return err
		}
	}

	if len(accs.LikedArtists) > 0 {
		_log.
			AddField("accountsCount", strconv.Itoa(len(accs.LikedArtists))).
			Info("Syncing liked artists...")
		if err := syncLikedArtists(ctx, accs.LikedArtists); err != nil {
			return err
		}
	}

	if len(accs.LikedTracks) > 0 {
		_log.
			AddField("accountsCount", strconv.Itoa(len(accs.LikedTracks))).
			Info("Syncing liked tracks...")
		if err := syncLikedTracks(ctx, accs.LikedTracks); err != nil {
			return err
		}
	}

	// if len(accs.Playlists) > 0 {
	// 	_log.
	// 		AddField("accountsCount", strconv.Itoa(len(accs.Playlists))).
	// 		Info("Syncing playlists...")
	// 	if err := syncPlaylists(ctx, accs.Playlists); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
