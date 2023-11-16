package streaming

import (
	"context"
	"net/url"
)

// Example: "Spotify".
type ServiceName string

func (r ServiceName) String() string {
	return string(r)
}

func (r *ServiceName) FromString(val string) {
	conv := ServiceName(val)
	*r = conv
}

// Example: artist ID on Spotify.
type ServiceEntityID string

func (r *ServiceEntityID) FromString(str string) {
	*r = ServiceEntityID(str)
}

func (r ServiceEntityID) String() string {
	return string(r)
}

type (
	// Example: Spotify.
	Service interface {
		// Provides Database for interacting with various things.
		//
		// Called once, after starting the program.
		Boot(Database) error

		// Unique name.
		Name() ServiceName

		// Get database.
		Database() Database

		// Assign actions for account.
		AssignAccountActions(Account) (AccountActions, error)

		// Get account-independent actions.
		//
		// Example: take one of the user accounts on streaming, and perform this actions from it.
		Actions() (ServiceActions, error)
	}

	// Entity.
	ServiceEntity interface {
		// Example: Spotify.
		ServiceName() ServiceName

		// Example: Spotify artist ID.
		ID() ServiceEntityID

		// Example: Spotify artist name.
		Name() string
	}

	// Artist.
	ServiceArtist interface {
		ServiceEntity

		// Albums names (oldest first).
		OldestAlbumsNames(context.Context) ([20]string, error)

		// Singles names (oldest first).
		OldestSinglesNames(context.Context) ([20]string, error)
	}

	// Album.
	ServiceAlbum interface {
		ServiceEntity

		// https://en.wikipedia.org/wiki/Universal_Product_Code
		UPC() *string

		// https://en.wikipedia.org/wiki/International_Article_Number
		EAN() *string

		Artists() []ServiceArtist

		TrackCount() int

		// Release year.
		Year() int

		// Cover (prefer 100x100).
		CoverURL() *url.URL
	}

	// Track.
	ServiceTrack interface {
		ServiceEntity

		// https://en.wikipedia.org/wiki/International_Standard_Recording_Code
		ISRC() *string

		Artists() []ServiceArtist

		Album() (ServiceAlbum, error)

		// Length of track in ms.
		LengthMs() int

		// Release year.
		Year() int

		// Cover (prefer 100x100).
		CoverURL() *url.URL
	}

	// Playlist created by user.
	ServicePlaylist interface {
		ServiceEntity

		FromAccount() Account

		// Should be nil only if service doesn't support set descriptions.
		Description() *string

		// Playlist tracks.
		Tracks(context.Context) ([]ServiceTrack, error)

		// Is playlist visible?
		//
		// Can return nil only if service doesn't support vis.
		IsVisible() *bool

		// Rename playlist.
		Rename(context.Context, string) error

		// Set playlist description.
		SetDescription(context.Context, string) error

		// Add tracks to playlist.
		AddTracks(context.Context, []ServiceEntityID) error

		// Remove tracks from playlist.
		RemoveTracks(context.Context, []ServiceEntityID) error

		// Set playlist visibility. True = visible. If service doesnt support vis, do nothing.
		SetIsVisible(context.Context, bool) error
	}
)
