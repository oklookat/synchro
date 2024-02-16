package shared

import (
	"context"
	"strconv"
	"time"
)

// Example: linked artist entity ID from DB.
type EntityID uint64

func (e EntityID) String() string {
	return strconv.FormatUint(uint64(e), 10)
}

func (e *EntityID) FromString(val string) error {
	conv, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return err
	}
	*e = EntityID(conv)
	return err
}

type (
	// Remote repository.
	RemoteRepository interface {
		// Unique ID.
		ID() string

		// Remote enabled? Not enabled remotes will be excluded from sync/linker/etc.
		Enabled() bool

		SetEnabled(bool) error

		// Remote name.
		Name() RemoteName

		// Alias: any text specified by the user so that he can distinguish one account from another.
		//
		// Auth: any auth like json tokens.
		//
		// After creating the account, the AssignActions method will be called in the remote.
		CreateAccount(alias string, auth string) (Account, error)

		// All accounts in remote.
		Accounts(ctx context.Context) ([]Account, error)

		// Get account by ID. Returns nil, nil if account not found.
		Account(id string) (Account, error)

		// Global remote actions.
		Actions() (RemoteActions, error)
	}

	// Remote account.
	Account interface {
		// Unique ID for account.
		ID() string

		// Account remote.
		RemoteName() RemoteName

		// Any text specified by the user so that he can distinguish one account from another.
		Alias() string

		// Set account alias.
		SetAlias(string) error

		// Account auth data like json tokens.
		Auth() string

		// Set account auth.
		SetAuth(string) error

		// Actions from remote account.
		Actions() (AccountActions, error)

		// Account settings.
		Settings() (AccountSettings, error)

		// Time when account was added.
		AddedAt() time.Time

		// Delete account from repository.
		Delete() error

		Repository() (RemoteRepository, error)
	}

	// Account settings.
	AccountSettings interface {
		// Liked albums synchronization.
		LikedAlbums() SynchronizationSettings

		// Liked artists synchronization.
		LikedArtists() SynchronizationSettings

		// Liked tracks synchronization.
		LikedTracks() SynchronizationSettings

		// Playlists synchronization.
		Playlists() SynchronizationSettings

		Playlist(playlistID EntityID) (PlaylistSyncSettings, error)
	}

	PlaylistSyncSettings interface {
		Name() SynchronizationSettings
		Description() SynchronizationSettings
		Visibility() SynchronizationSettings
		Tracks() SynchronizationSettings
	}

	// Remote account synchronization settings.
	SynchronizationSettings interface {
		// Synchronization enabled?
		Synchronize() bool

		// Enable/disable synchronization.
		SetSynchronize(bool) error

		// Last account synchronization (nanoseconds).
		LastSynchronization() time.Time

		// Set last account synchronization (nanoseconds).
		//
		// Do not use this method in the UI and similar places. It is only needed for the synchronizer.
		SetLastSynchronization(time.Time) error
	}
)
