package commander

import (
	"context"
	"strconv"

	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
)

func AccountByID(id string) (*Account, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	acc, err := repository.AccountByID(uid)
	return newAccount(acc), err
}

func Accounts(remoteName string) (*AccountSlice, error) {
	repo, ok := repository.Remotes[shared.RemoteName(remoteName)]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(_packageName, shared.RemoteName(remoteName))
	}
	accs, err := repo.Repository().Accounts(context.Background())
	return newAccountSlice(accs), err
}

func DeleteAccount(id string) error {
	return execTask(0, func(ctx context.Context) error {
		acc, err := accountByID(id)
		if err != nil {
			return err
		}
		return acc.Delete()
	})
}

func newAccount(from shared.Account) *Account {
	return &Account{
		self: from,
	}
}

type Account struct {
	self shared.Account
}

func (e *Account) ID() string {
	return e.self.ID()
}

func (e *Account) Alias() string {
	return e.self.Alias()
}

func (e *Account) SetAlias(val string) error {
	return e.self.SetAlias(val)
}

func (e *Account) Auth() string {
	return e.self.Auth()
}

func (e *Account) SetAuth(val string) error {
	return execTask(0, func(ctx context.Context) error {
		return e.self.SetAuth(val)
	})
}

func (e *Account) AddedAt() string {
	return e.self.AddedAt().String()
}

func (e *Account) Settings() (*AccountSettings, error) {
	setts, err := e.self.Settings()
	return &AccountSettings{
		self: setts,
	}, err
}

func newAccountSlice(accs []shared.Account) *AccountSlice {
	converted := make([]Account, len(accs))
	for i := range converted {
		converted[i] = *newAccount(accs[i])
	}
	return &AccountSlice{
		wrap: &wrappedSlice[Account]{
			items: converted,
		},
	}
}

type AccountSlice struct {
	wrap *wrappedSlice[Account]
}

func (e *AccountSlice) Item(i int) *Account {
	return e.wrap.Item(i)
}
func (e *AccountSlice) Len() int {
	return e.wrap.Len()
}

type AccountSettings struct {
	self shared.AccountSettings
}

func (e *AccountSettings) LikedAlbums() *SyncSetting {
	return &SyncSetting{self: e.self.LikedAlbums()}
}

func (e *AccountSettings) LikedArtists() *SyncSetting {
	return &SyncSetting{self: e.self.LikedArtists()}
}

func (e *AccountSettings) LikedTracks() *SyncSetting {
	return &SyncSetting{self: e.self.LikedTracks()}
}

func (e *AccountSettings) Playlists() *SyncSetting {
	return &SyncSetting{self: e.self.Playlists()}
}

func (e *AccountSettings) Playlist(playlistID string) (*PlaylistSyncSettings, error) {
	id, err := strconv.ParseUint(playlistID, 10, 64)
	if err != nil {
		return nil, err
	}
	setts, err := e.self.Playlist(shared.EntityID(id))
	return &PlaylistSyncSettings{
		self: setts,
	}, err
}

type PlaylistSyncSettings struct {
	self shared.PlaylistSyncSettings
}

func (e *PlaylistSyncSettings) Name() *SyncSetting {
	return &SyncSetting{self: e.self.Name()}
}

func (e *PlaylistSyncSettings) Description() *SyncSetting {
	return &SyncSetting{self: e.self.Description()}
}

func (e *PlaylistSyncSettings) Visibility() *SyncSetting {
	return &SyncSetting{self: e.self.Visibility()}
}

func (e *PlaylistSyncSettings) Tracks() *SyncSetting {
	return &SyncSetting{self: e.self.Tracks()}
}

type SyncSetting struct {
	self shared.SynchronizationSettings
}

func (e *SyncSetting) Synchronize() bool {
	return e.self.Synchronize()
}

func (e *SyncSetting) SetSynchronize(val bool) error {
	return execTask(0, func(ctx context.Context) error {
		return e.self.SetSynchronize(val)
	})
}

func (e *SyncSetting) LastSynchronization() string {
	return e.self.LastSynchronization().String()
}
