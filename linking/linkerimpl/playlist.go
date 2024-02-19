package linkerimpl

// import (
// 	"context"

// 	"github.com/oklookat/synchro/linking/linker"
// 	"github.com/oklookat/synchro/repository"
// 	"github.com/oklookat/synchro/shared"
// )

// func NewPlaylists(accounts []shared.Account) (*linker.Dynamic, error) {
// 	if len(accounts) == 0 {
// 		_log.Error("NewPlaylists len(accounts) == 0")
// 		return nil, shared.NewErrNoAvailableRemotes(_packageName)
// 	}
// 	converted := map[shared.RemoteName]linker.RemoteDynamic{}
// 	for i, acc := range accounts {
// 		converted[shared.RemoteName(acc.ID())] = PlaylistsRemote{account: accounts[i]}
// 	}
// 	return linker.NewDynamic(NewPlaylistsRepository(), converted), nil
// }

// func NewPlaylistsRepository() linker.RepositoryDynamic {
// 	return repository.PlaylistEntity
// }

// type PlaylistsRemote struct {
// 	account shared.Account
// }

// func (e PlaylistsRemote) Name() shared.RemoteName {
// 	return shared.RemoteName(e.account.ID())
// }

// func (e PlaylistsRemote) Create(ctx context.Context, name string) (linker.RemoteEntity, error) {
// 	actions, err := e.account.Actions()
// 	if err != nil {
// 		return nil, err
// 	}
// 	action := actions.Playlist()
// 	created, err := action.Create(ctx, name, false, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewPlaylistRemoteEntity(created, e.account.ID()), err
// }

// func (e PlaylistsRemote) Linkables() linker.LinkablesDynamic {
// 	real, ok := e.account.(*repository.Account)
// 	if !ok {
// 		_log.Error("real, ok := e.repo.(*repository.Account)")
// 		return nil
// 	}
// 	return repository.NewLinkablePlaylist(real)
// }

// func NewPlaylistRemoteEntity(real shared.RemotePlaylist, fromAccountID string) *PlaylistRemoteEntity {
// 	if real == nil {
// 		return nil
// 	}
// 	return &PlaylistRemoteEntity{
// 		Real:          real,
// 		FromAccountID: fromAccountID,
// 	}
// }

// type PlaylistRemoteEntity struct {
// 	FromAccountID string
// 	Real          shared.RemotePlaylist
// }

// // Example: Spotify.
// func (e PlaylistRemoteEntity) RemoteName() shared.RemoteName {
// 	return shared.RemoteName(e.FromAccountID)
// }

// // Example: Spotify artist ID.
// func (e PlaylistRemoteEntity) ID() shared.RemoteID {
// 	return e.Real.ID()
// }

// // Example: Spotify artist name.
// func (e PlaylistRemoteEntity) Name() string {
// 	return e.Real.Name()
// }
