package snapshot

import (
	"context"
	"time"

	"github.com/oklookat/synchro/shared"
)

type SnapshotsFilterAuto string

func (e *SnapshotsFilterAuto) FromString(val string) {
	*e = SnapshotsFilterAuto(val)
}

const (
	// Get only auto snapshots.
	SnapshotsFilterAutoAuto SnapshotsFilterAuto = "auto"

	// Get only manual snapshots.
	SnapshotsFilterAutoManual SnapshotsFilterAuto = "manual"

	// Get all.
	SnapshotsFilterAutoAll SnapshotsFilterAuto = "all"
)

type (
	// Manages snapshots for specific remote account.
	Snapshotter interface {
		Create(in shared.RemoteName, alias string, auto bool) (Snapshot, error)

		// If remote name empty - get snapshots from all remotes.
		Snapshots(
			remoteName shared.RemoteName,
			filterAuto SnapshotsFilterAuto,
		) ([]Snapshot, error)

		Snapshot(id shared.RepositoryID) (Snapshot, error)

		// Delete oldest auto snapshots in remote, if max count reached.
		//
		// I.e: 50 auto snapshots, max = 30.
		// Result: 30 new ones should remain.
		DeleteOldestAuto(in shared.RemoteName, max int) error
	}

	Snapshot interface {
		// Unique ID.
		ID() shared.RepositoryID

		// Snapshot belongs to this remote.
		RemoteName() shared.RemoteName

		// Automatic snapshot?
		Auto() bool

		// Just alias for user.
		Alias() string

		// Set alias.
		SetAlias(string) error

		/*
			About "restoreable":
			We use this field to separate the null snapshot data from the full snapshot data.
			For example, we have a snapshot where there are no liked albums. There are two options here:
			1. The snapshot was created with liked albums,
				but there were no liked albums in the account.
			2. A snapshot was created excluding liked albums.

			In the first case, restoring from a snapshot will delete all the liked albums from the account.
			In the second case, the liked albums will not be affected.

			Default value of "resotreable" must be "false".

			After calling Set method, also set "restoreable" to "true".
		*/

		// Liked albums.
		LikedAlbumsRestoreable() bool
		LikedAlbumsCount() (int, error)
		LikedAlbums() ([]shared.RemoteID, error)
		SetLikedAlbums([]shared.RemoteID) error
		RestoreLikedAlbums(ctx context.Context, merge bool, action shared.LikedActions) error

		// Liked artists.
		LikedArtistsRestoreable() bool
		LikedArtistsCount() (int, error)
		LikedArtists() ([]shared.RemoteID, error)
		SetLikedArtists([]shared.RemoteID) error
		RestoreLikedArtists(ctx context.Context, merge bool, action shared.LikedActions) error

		// Liked tracks.
		LikedTracksRestoreable() bool
		LikedTracksCount() (int, error)
		LikedTracks() ([]shared.RemoteID, error)
		SetLikedTracks([]shared.RemoteID) error
		RestoreLikedTracks(ctx context.Context, merge bool, action shared.LikedActions) error

		// Playlists.
		PlaylistsRestoreable() bool
		PlaylistsCount() (int, error)

		// Get playlist snapshot by ID.
		Playlist(string) (Playlist, error)
		Playlists() ([]Playlist, error)
		AddPlaylist(name string,
			description *string,
			isVisible bool,
			tracks []shared.RemoteID) (Playlist, error)
		RestorePlaylists(ctx context.Context, merge bool, action shared.PlaylistActions) error

		// Snapshot creation date.
		CreatedAt() time.Time

		// Delete snapshot.
		Delete() error
	}

	Playlist interface {
		// Unique repository playlist ID.
		ID() shared.RepositoryID

		Name() string

		IsVisible() bool

		// Can be nil only if remote doesn't support descriptions.
		Description() *string

		Tracks() ([]shared.RemoteID, error)

		Restore(context.Context, shared.PlaylistActions) error

		// Delete playlist from snapshot.
		Delete() error

		// Snapshot creation date.
		CreatedAt() time.Time
	}
)
