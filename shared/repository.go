package shared

import (
	"context"
	"time"
)

// Example: ID in database.
type RepositoryID string

func (e RepositoryID) String() string {
	return string(e)
}

// Example: linked artist entity ID from DB.
type EntityID RepositoryID

func (e EntityID) String() string {
	return string(RepositoryID(e))
}

type (
	// Remote repository.
	RemoteRepository interface {
		// Unique ID.
		ID() RepositoryID

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
		Account(id RepositoryID) (Account, error)

		// Global remote actions.
		Actions() (RemoteActions, error)
	}

	// Remote account.
	Account interface {
		// Unique ID for account.
		ID() RepositoryID

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
