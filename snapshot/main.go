package snapshot

import (
	"context"
	"errors"
	"time"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

type CreateSettings struct {
	Service       streaming.ServiceName    `json:"service"`
	Actions       streaming.AccountActions `json:"actions"`
	Artists       bool                     `json:"artists"`
	Albums        bool                     `json:"albums"`
	Tracks        bool                     `json:"tracks"`
	Playlists     bool                     `json:"playlists"`
	UserPlaylists bool                     `json:"userPlaylists"`
}

func Create(ctx context.Context, settings CreateSettings) (*Full, error) {
	if shared.IsNil(settings.Actions) {
		return nil, errors.New("nil account actions")
	}

	result := &Full{
		ServiceName: settings.Service,
		Version:     1,
		CreatedAt:   time.Now(),
	}
	act := settings.Actions

	if settings.Albums {
		ids, err := getLikedIds(ctx, act.LikedAlbums())
		if err != nil {
			return nil, err
		}
		result.Albums = ids
	}

	if settings.Artists {
		ids, err := getLikedIds(ctx, act.LikedArtists())
		if err != nil {
			return nil, err
		}
		result.Artists = ids
	}

	if settings.Tracks {
		ids, err := getLikedIds(ctx, act.LikedTracks())
		if err != nil {
			return nil, err
		}
		result.Tracks = ids
	}

	if settings.UserPlaylists {
		pls, err := act.Playlist().MyPlaylists(ctx)
		if err != nil {
			return nil, err
		}

		snapPls := make([]UserPlaylist, len(pls))
		for i := range snapPls {
			uPl := UserPlaylist{
				Name:        pls[i].Name(),
				Description: pls[i].Description(),
				IsVisible:   pls[i].IsVisible(),
			}

			plsTracks, err := pls[i].Tracks(ctx)
			if err != nil {
				return nil, err
			}

			uPl.Tracks = make([]streaming.ServiceEntityID, len(plsTracks))
			for i2 := range uPl.Tracks {
				uPl.Tracks[i2] = plsTracks[i2].ID()
			}
		}
	}

	return result, nil
}

func getLikedIds(ctx context.Context, acts streaming.LikedActions) ([]streaming.ServiceEntityID, error) {
	likes, err := acts.Liked(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]streaming.ServiceEntityID, len(likes))
	for i := range result {
		result[i] = likes[i].ID()
	}
	return result, err
}

type UserPlaylist struct {
	Name        string                      `json:"name"`
	Description *string                     `json:"description"`
	IsVisible   *bool                       `json:"isVisible"`
	Tracks      []streaming.ServiceEntityID `json:"tracks"`
}
