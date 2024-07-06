package shared

import (
	"context"
	"net/url"
)

// Must be implemented by a remote.
type (
	// Example: music streaming service.
	Remote interface {
		// Provides repository for interacting with various things.
		//
		// Called once, after starting the program.
		Boot(RemoteRepository) error

		// Unique remote name.
		Name() RemoteName

		// Get repository.
		Repository() RemoteRepository

		// Assign actions for account.
		AssignAccountActions(Account) (AccountActions, error)

		// Get actions for remote.
		//
		// Example: take one of the accounts, and perform this actions from it.
		Actions() (RemoteActions, error)

		// Get url to entity.
		EntityURL(etype EntityType, id RemoteID) url.URL
	}

	// Remote entity.
	RemoteEntity interface {
		// Example: Spotify.
		RemoteName() RemoteName

		// Example: Spotify artist ID.
		ID() RemoteID

		// Example: Spotify artist name.
		Name() string
	}

	// Artist from remote.
	RemoteArtist interface {
		RemoteEntity

		// Albums names (oldest first).
		OldestAlbumsNames(ctx context.Context) ([20]string, error)

		// Singles names (oldest first).
		OldestSinglesNames(ctx context.Context) ([20]string, error)
	}

	// Album from remote.
	RemoteAlbum interface {
		RemoteEntity

		// https://en.wikipedia.org/wiki/Universal_Product_Code
		UPC() *string

		// https://en.wikipedia.org/wiki/International_Article_Number
		EAN() *string

		Artists() []RemoteArtist

		TrackCount() int

		// Release year.
		Year() int

		// Cover (prefer 100x100).
		CoverURL() *url.URL
	}

	// Track from remote.
	RemoteTrack interface {
		RemoteEntity

		// https://en.wikipedia.org/wiki/International_Standard_Recording_Code
		ISRC() *string

		Artists() []RemoteArtist

		Album() (RemoteAlbum, error)

		// Length of track in ms.
		LengthMs() int

		// Release year.
		Year() int

		// Cover (prefer 100x100).
		CoverURL() *url.URL
	}

	// User playlist from remote.
	RemotePlaylist interface {
		RemoteEntity

		FromAccount() Account

		// Should be nil only if remote doesn't support set descriptions.
		Description() *string

		// Playlist tracks.
		Tracks(context.Context) ([]RemoteTrack, error)

		// Rename playlist.
		Rename(context.Context, string) error

		// Set playlist description.
		SetDescription(context.Context, string) error

		// Add tracks to playlist.
		AddTracks(context.Context, []RemoteID) error

		// Remove tracks from playlist.
		RemoveTracks(context.Context, []RemoteID) error

		// Is playlist visible?
		//
		// Can return error only if remote doesn't support vis.
		IsVisible() (bool, error)

		// Set playlist visibility. True = visible.
		SetIsVisible(context.Context, bool) error
	}

	// Actions for remote account.
	//
	// Only ErrNotImplemented can be returned as an error.
	// But you must implement at least LikedArtists().
	AccountActions interface {
		LikedAlbums() LikedActions
		LikedArtists() LikedActions
		LikedTracks() LikedActions
		Playlist() PlaylistActions
	}

	// General actions for remote.
	RemoteActions interface {
		// Get artist by ID.
		Artist(context.Context, RemoteID) (RemoteArtist, error)

		// Get track by ID.
		Track(context.Context, RemoteID) (RemoteTrack, error)

		// Get album by ID.
		Album(context.Context, RemoteID) (RemoteAlbum, error)

		// Search albums.
		SearchAlbums(context.Context, RemoteAlbum) ([10]RemoteAlbum, error)

		// Search artists.
		SearchArtists(context.Context, RemoteArtist) ([10]RemoteArtist, error)

		// Search tracks.
		SearchTracks(context.Context, RemoteTrack) ([10]RemoteTrack, error)
	}

	// Actions for entities created by the remote.
	//
	// Example: artist, track, album.
	LikedActions interface {
		// Examples: get liked artists.
		Liked(context.Context) ([]RemoteEntity, error)

		// Examples: like artists.
		Like(ctx context.Context, ids []RemoteID) error

		// Examples: unlike artists.
		Unlike(ctx context.Context, ids []RemoteID) error
	}

	// Playlist actions.
	PlaylistActions interface {
		// Get user playlists.
		MyPlaylists(context.Context) ([]RemotePlaylist, error)

		// Create playlist.
		Create(ctx context.Context, name string, isVisible bool, description *string) (RemotePlaylist, error)

		// Delete my playlists.
		Delete(ctx context.Context, entities []RemoteID) error

		// Get playlist by ID.
		Playlist(context.Context, RemoteID) (RemotePlaylist, error)
	}
)

type EntityType string

func (e EntityType) String() string {
	return string(e)
}

func (e *EntityType) FromString(val string) {
	conv := EntityType(val)
	*e = conv
}

const (
	EntityTypeAlbum    EntityType = "album"
	EntityTypeArtist   EntityType = "artist"
	EntityTypeTrack    EntityType = "track"
	EntityTypePlaylist EntityType = "playlist"
)
