package commander

import (
	"context"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
)

// Remote active and ready for sync & tools?
func RemoteEnabled(remoteName string) bool {
	var remName shared.RemoteName
	remName.FromString(remoteName)
	rem, ok := _remotes[remName]
	if !ok {
		return false
	}
	return rem.Repository().Enabled()
}

func SetRemoteEnabled(remoteName string, val bool) error {
	return execTask(0, func(ctx context.Context) error {
		var remName shared.RemoteName
		remName.FromString(remoteName)
		rem, ok := _remotes[remName]
		if !ok {
			return shared.NewErrRemoteNotFound(_packageName, remName)
		}
		return rem.Repository().SetEnabled(val)
	})
}

func UnlikeLikedAlbums(accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		acc, err := accountByID(accountID)
		if err != nil {
			return err
		}
		actions, err := acc.Actions()
		if err != nil {
			return err
		}
		return unlikeLiked(ctx, actions.LikedAlbums())
	})
}

func UnlikeLikedArtists(accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		acc, err := accountByID(accountID)
		if err != nil {
			return err
		}
		actions, err := acc.Actions()
		if err != nil {
			return err
		}
		return unlikeLiked(ctx, actions.LikedArtists())
	})
}

func UnlikeLikedTracks(accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		acc, err := accountByID(accountID)
		if err != nil {
			return err
		}
		actions, err := acc.Actions()
		if err != nil {
			return err
		}
		return unlikeLiked(ctx, actions.LikedTracks())
	})
}

func DeleteAllPlaylists(accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		acc, err := accountByID(accountID)
		if err != nil {
			return err
		}
		actions, err := acc.Actions()
		if err != nil {
			return err
		}

		playlists, err := actions.Playlist().MyPlaylists(ctx)
		if err != nil {
			return err
		}

		var ids shared.RemoteIDSlice[shared.RemotePlaylist]
		ids.FromMap(playlists)

		return actions.Playlist().Delete(ctx, ids)
	})
}

func unlikeLiked(ctx context.Context, actions shared.LikedActions) error {
	entities, err := actions.Liked(ctx)
	if err != nil {
		return err
	}
	if len(entities) == 0 {
		return nil
	}
	var conv shared.RemoteIDSlice[shared.RemoteEntity]
	conv.FromMap(entities)
	return actions.Unlike(ctx, conv)
}

func transferIds(
	ctx context.Context,
	lnk *linker.Static,

	getEntity func(ctx context.Context, id shared.RemoteID) (shared.RemoteEntity, error),
	entitiesIds []shared.RemoteID,

	to shared.RemoteName,
) ([]shared.RemoteID, error) {
	if len(entitiesIds) == 0 {
		return nil, nil
	}

	var entities []shared.RemoteEntity
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

	var targetIds []shared.RemoteID
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
