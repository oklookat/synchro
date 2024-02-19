package lisybridge

import (
	"context"

	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncerimpl"
)

var (
	_syncablesLikedTracks = repository.TrackSyncable
)

func syncLikedTracks(ctx context.Context, accounts map[shared.RepositoryID]*fullAccount) error {
	lnk, err := linkerimpl.NewTracks()
	if err != nil {
		return err
	}

	var likeableAccounts []*syncerimpl.LikeableAccount
	for _, acc := range accounts {
		acts := acc.Actions.LikedTracks()
		likes, err := acts.Liked(ctx)
		if err != nil {
			return err
		}

		if err := onGotLikedTracks(ctx, acc.Account, likes); err != nil {
			return err
		}

		likesConverted, err := linkStatic(ctx, lnk, likes)
		if err != nil {
			return err
		}

		setts := acc.Settings.LikedTracks()
		wrapAcc := syncerimpl.NewLikeableAccount(
			_syncablesLikedTracks,
			lnk,
			acc.RemoteName,
			likesConverted,
			setts.LastSynchronization(),
			setts.SetLastSynchronization,
			acts,
		)
		likeableAccounts = append(likeableAccounts, wrapAcc)
	}

	defer _syncablesLikedTracks.DeleteUnsynced()
	return sync2stages(func() error {
		for _, acc := range likeableAccounts {
			if err := acc.Start(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
