package snapshot

import (
	"context"
	"errors"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

type RecoverSettings struct {
	target        streaming.AccountActions
	Albums        bool `json:"albums"`
	Artists       bool `json:"artists"`
	Tracks        bool `json:"tracks"`
	Playlists     bool `json:"playlists"`
	UserPlaylists bool `json:"userPlaylists"`
}

type Full struct {
	ServiceName   streaming.ServiceName       `json:"serviceName"`
	Version       int                         `json:"version"`
	Albums        []streaming.ServiceEntityID `json:"albums"`
	Artists       []streaming.ServiceEntityID `json:"artists"`
	Tracks        []streaming.ServiceEntityID `json:"tracks"`
	UserPlaylists []UserPlaylist              `json:"userPlaylists"`
	CreatedAt     time.Time                   `json:"createdAt"`
}

// Recover snapshot.
func (f Full) Recover(ctx context.Context, setts RecoverSettings) error {
	if shared.IsNil(setts.target) {
		return errors.New("nil account actions")
	}

	if len(f.Albums) > 0 {
		if err := f.recoverLiked(ctx, setts.target.LikedAlbums(), f.Albums); err != nil {
			return err
		}
	}

	if len(f.Artists) > 0 {
		if err := f.recoverLiked(ctx, setts.target.LikedArtists(), f.Artists); err != nil {
			return err
		}
	}

	if len(f.Tracks) > 0 {
		if err := f.recoverLiked(ctx, setts.target.LikedTracks(), f.Tracks); err != nil {
			return err
		}
	}

	if len(f.UserPlaylists) > 0 {
		plAct := setts.target.Playlist()
		for _, up := range f.UserPlaylists {
			isVisible := false
			if up.IsVisible != nil {
				isVisible = *up.IsVisible
			}
			newPl, err := plAct.Create(ctx, up.Name, isVisible, up.Description)
			if err != nil {
				return err
			}
			if err = newPl.AddTracks(ctx, up.Tracks); err != nil {
				return err
			}
		}
	}

	return nil
}

// Convert snapshot to snapshot in another streaming.
func (f Full) CrossShot(ctx context.Context, origin streaming.Service, target streaming.Service) (*Full, error) {
	if origin.Name() == target.Name() {
		return nil, errors.New("origin and target have same name")
	}

	result := &Full{
		ServiceName: target.Name(),
		Version:     1,
		CreatedAt:   time.Now(),
	}

	originActs, err := origin.Actions()
	if err != nil {
		return nil, err
	}

	if len(f.Albums) > 0 {
		lnk, err := linkerimpl.NewAlbums()
		if err != nil {
			return nil, err
		}
		getWrap := func(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceEntity, error) {
			return originActs.Album(ctx, id)
		}
		converted, err := transferIds(ctx, lnk, getWrap, f.Albums, target.Name())
		if err != nil {
			return nil, err
		}
		result.Albums = converted
	}

	if len(f.Artists) > 0 {
		lnk, err := linkerimpl.NewArtists()
		if err != nil {
			return nil, err
		}
		getWrap := func(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceEntity, error) {
			return originActs.Artist(ctx, id)
		}
		converted, err := transferIds(ctx, lnk, getWrap, f.Artists, target.Name())
		if err != nil {
			return nil, err
		}
		result.Artists = converted
	}

	lnkTracks, err := linkerimpl.NewTracks()
	if err != nil {
		return nil, err
	}
	getTrackWrap := func(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceEntity, error) {
		return originActs.Track(ctx, id)
	}

	if len(f.Tracks) > 0 {
		converted, err := transferIds(ctx, lnkTracks, getTrackWrap, f.Tracks, target.Name())
		if err != nil {
			return nil, err
		}
		result.Tracks = converted
	}

	if len(f.UserPlaylists) > 0 {
		for i := range f.UserPlaylists {
			converted, err := transferIds(ctx, lnkTracks, getTrackWrap,
				f.UserPlaylists[i].Tracks, target.Name())
			if err != nil {
				return nil, err
			}
			result.UserPlaylists = append(result.UserPlaylists, UserPlaylist{
				Name:        f.UserPlaylists[i].Name,
				Description: f.UserPlaylists[i].Description,
				IsVisible:   f.UserPlaylists[i].IsVisible,
				Tracks:      converted,
			})
		}
	}

	return result, err
}

func (f Full) recoverLiked(ctx context.Context, act streaming.LikedActions, ids []streaming.ServiceEntityID) error {
	return act.Like(ctx, ids)
}

func transferIds(
	ctx context.Context,
	lnk *linker.Static,

	getEntity func(ctx context.Context, id streaming.ServiceEntityID) (streaming.ServiceEntity, error),
	entitiesIds []streaming.ServiceEntityID,

	to streaming.ServiceName,
) ([]streaming.ServiceEntityID, error) {
	if len(entitiesIds) == 0 {
		return nil, nil
	}

	var entities []streaming.ServiceEntity
	for _, id := range entitiesIds {
		ent, err := getEntity(ctx, id)
		if err != nil {
			return nil, err
		}
		entities = append(entities, ent)
	}

	var originLinked []linker.Linked
	for _, ent := range entities {
		result, err := lnk.FromRemote(ctx, ent)
		if err != nil {
			return nil, err
		}
		if shared.IsNil(result.Linked) {
			continue
		}
		originLinked = append(originLinked, result.Linked)
	}

	var targetIds []streaming.ServiceEntityID
	for _, link := range originLinked {
		result, err := lnk.ToRemote(ctx, link.EntityID(), to)
		if err != nil {
			return nil, err
		}
		// Missing.
		if shared.IsNil(result.Linked) || result.Linked.RemoteID() == nil {
			continue
		}
		targetIds = append(targetIds, *result.Linked.RemoteID())
	}

	return targetIds, nil
}
