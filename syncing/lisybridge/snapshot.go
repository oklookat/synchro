package lisybridge

import (
	"context"
	"fmt"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/snapshot"
)

type snapshotHooker struct {
	// [remoteName]Snapshot
	snaps map[shared.RemoteName]snapshot.Snapshot

	cfg *config.Snapshots
}

func (h *snapshotHooker) onLikedAlbums(ctx context.Context, account shared.Account, liked map[shared.RemoteID]shared.RemoteEntity) error {
	repo, err := account.Repository()
	remName := repo.Name()
	if err != nil {
		return err
	}
	if err := h.manageSnaps(remName, account); err != nil {
		return err
	}
	if h.cfg.CreateWhenSyncing {
		snap := h.snaps[remName]
		var likes shared.RemoteIDSlice[shared.RemoteEntity]
		likes.FromMap(liked)
		return snap.SetLikedAlbums(likes)
	}
	return nil
}

func (h *snapshotHooker) onLikedArtists(ctx context.Context, account shared.Account, liked map[shared.RemoteID]shared.RemoteEntity) error {
	repo, err := account.Repository()
	remName := repo.Name()
	if err != nil {
		return err
	}
	if err := h.manageSnaps(remName, account); err != nil {
		return err
	}
	if h.cfg.CreateWhenSyncing {
		snap := h.snaps[remName]
		var likes shared.RemoteIDSlice[shared.RemoteEntity]
		likes.FromMap(liked)
		return snap.SetLikedArtists(likes)
	}
	return nil
}

func (h *snapshotHooker) onLikedTracks(ctx context.Context, account shared.Account, liked map[shared.RemoteID]shared.RemoteEntity) error {
	repo, err := account.Repository()
	remName := repo.Name()
	if err != nil {
		return err
	}
	if err := h.manageSnaps(remName, account); err != nil {
		return err
	}
	if h.cfg.CreateWhenSyncing {
		snap := h.snaps[remName]
		var likes shared.RemoteIDSlice[shared.RemoteEntity]
		likes.FromMap(liked)
		return snap.SetLikedTracks(likes)
	}
	return nil
}

func (h *snapshotHooker) onPlaylists(ctx context.Context, account shared.Account, playlists map[shared.RemoteID]shared.RemotePlaylist) error {
	repo, err := account.Repository()
	remName := repo.Name()
	if err != nil {
		return err
	}
	if err := h.manageSnaps(remName, account); err != nil {
		return err
	}
	if !h.cfg.CreateWhenSyncing {
		return err
	}
	snap := h.snaps[remName]
	for _, pl := range playlists {
		tracks, err := pl.Tracks(context.Background())
		if err != nil {
			return err
		}
		var conv shared.RemoteIDSlice[shared.RemoteTrack]
		conv.FromMap(tracks)
		vis, _ := pl.IsVisible()
		_, err = snap.AddPlaylist(pl.Name(), pl.Description(), vis, conv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *snapshotHooker) cache() error {
	if h.snaps == nil {
		h.snaps = map[shared.RemoteName]snapshot.Snapshot{}
	}
	if h.cfg == nil {
		cfg, err := config.Get[config.Snapshots](config.KeySnapshots)
		if err != nil {
			return err
		}
		h.cfg = cfg
	}
	return nil
}

func (h *snapshotHooker) manageSnaps(where shared.RemoteName, account shared.Account) error {
	h.cache()
	_, ok := h.snaps[where]
	if ok {
		return nil
	}
	if !h.cfg.CreateWhenSyncing {
		return nil
	}
	snap, err := repository.Snap.Create(where, fmt.Sprintf("(%s) %s", account.Alias(), shared.TimeNowStr()), true)
	if err != nil {
		return err
	}
	// For example in Android, Go gets time with (0 GMT?) timezone.
	// So we just use id instead of time.
	if err = snap.SetAlias(fmt.Sprintf("(%s) %s", snap.ID(), account.Alias())); err != nil {
		return err
	}
	if err := repository.Snap.DeleteOldestAuto(where, h.cfg.MaxAuto); err != nil {
		return err
	}
	h.snaps[where] = snap
	return nil
}
