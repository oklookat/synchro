package lisybridge

import (
	"context"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/shared"
)

// Convert static to linked for account.
func linkStatic(
	ctx context.Context,
	lnk *linker.Static,
	entities map[shared.RemoteID]shared.RemoteEntity,
) (map[shared.EntityID]shared.RemoteID, error) {
	result := make(map[shared.EntityID]shared.RemoteID, len(entities))
	for id := range entities {
		lnkResult, err := lnk.FromRemote(
			ctx,
			linkerimpl.NewRemoteEntity(entities[id]),
		)
		if err != nil {
			return nil, err
		}
		if shared.IsNil(lnkResult.Linked) || lnkResult.Linked.RemoteID() == nil {
			continue
		}
		result[lnkResult.Linked.EntityID()] = *lnkResult.Linked.RemoteID()
	}

	return result, nil
}

func sync2stages(executor func() error) error {
	_log.Info("Synchronizing something...")
	for x := 0; x < 2; x++ {
		if err := executor(); err != nil {
			return err
		}
	}
	return nil
}

func getAccountsForSync(ctx context.Context) (accountsForSync, error) {
	result := accountsForSync{
		LikedAlbums:  map[shared.RepositoryID]*fullAccount{},
		LikedArtists: map[shared.RepositoryID]*fullAccount{},
		LikedTracks:  map[shared.RepositoryID]*fullAccount{},
		Playlists:    map[shared.RepositoryID]*fullAccount{},
	}

	for _, remote := range _remotes {
		if !remote.Repository().Enabled() {
			continue
		}

		accounts, err := remote.Repository().Accounts(ctx)
		if err != nil {
			return result, err
		}

		for i := range accounts {
			acc := accounts[i]
			fullAcc, err := newFullAccount(acc)
			if err != nil {
				return result, err
			}
			if fullAcc.Settings.LikedAlbums().Synchronize() {
				result.LikedAlbums[acc.ID()] = fullAcc
			}
			if fullAcc.Settings.LikedArtists().Synchronize() {
				result.LikedArtists[acc.ID()] = fullAcc
			}
			if fullAcc.Settings.LikedTracks().Synchronize() {
				result.LikedTracks[acc.ID()] = fullAcc
			}
			if fullAcc.Settings.Playlists().Synchronize() {
				result.Playlists[acc.ID()] = fullAcc
			}
		}
	}

	return result, nil
}

// [account id]account
type accountsForSync struct {
	LikedAlbums,
	LikedArtists,
	LikedTracks,
	Playlists map[shared.RepositoryID]*fullAccount
}

func newFullAccount(acc shared.Account) (*fullAccount, error) {
	acts, err := acc.Actions()
	if err != nil {
		return nil, err
	}
	setts, err := acc.Settings()
	if err != nil {
		return nil, err
	}
	return &fullAccount{
		Account:         acc,
		RemoteName:      acc.RemoteName(),
		Actions:         acts,
		Settings:        setts,
		LinkedPlaylists: map[shared.EntityID]shared.RemotePlaylist{},
	}, nil
}

type fullAccount struct {
	Account    shared.Account
	RemoteName shared.RemoteName
	Actions    shared.AccountActions
	Settings   shared.AccountSettings

	// Available on playlists sync.
	LinkedPlaylists map[shared.EntityID]shared.RemotePlaylist
}
