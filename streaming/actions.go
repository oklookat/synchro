package streaming

import "context"

type (
	// Actions for service account.
	//
	// Only ErrNotImplemented can be returned as an error.
	AccountActions interface {
		LikedAlbums() LikedActions
		LikedArtists() LikedActions
		LikedTracks() LikedActions
		Playlist() PlaylistActions
	}

	// Account-independent actions.
	ServiceActions interface {
		// Get artist by ID.
		Artist(context.Context, ServiceEntityID) (ServiceArtist, error)

		// Get track by ID.
		Track(context.Context, ServiceEntityID) (ServiceTrack, error)

		// Get album by ID.
		Album(context.Context, ServiceEntityID) (ServiceAlbum, error)

		// Search albums.
		SearchAlbums(context.Context, ServiceAlbum) ([10]ServiceAlbum, error)

		// Search artists.
		SearchArtists(context.Context, ServiceArtist) ([10]ServiceArtist, error)

		// Search tracks.
		SearchTracks(context.Context, ServiceTrack) ([10]ServiceTrack, error)
	}

	// Actions for entities created by the service.
	//
	// Example: artist, track, album.
	LikedActions interface {
		// Examples: get liked artists.
		Liked(context.Context) ([]ServiceEntity, error)

		// Examples: like artists.
		Like(context.Context, []ServiceEntityID) error

		// Examples: unlike artists.
		Unlike(context.Context, []ServiceEntityID) error
	}

	// Playlist actions.
	PlaylistActions interface {
		// Get user created playlists.
		MyPlaylists(context.Context) ([]ServicePlaylist, error)

		// Get user playlist by ID.
		MyPlaylist(context.Context, ServiceEntityID) (ServicePlaylist, error)

		// Create playlist.
		Create(ctx context.Context, name string, isVisible bool, description *string) (ServicePlaylist, error)

		// Delete my playlists.
		Delete(context.Context, []ServiceEntityID) error
	}
)
