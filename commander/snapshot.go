package commander

import (
	"context"
	"errors"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/snapshot"
)

func NewConfigSnapshots() *ConfigSnapshots {
	cfg := &config.Snapshots{}
	if err := config.Get(cfg); err != nil {
		cfg.Default()
	}
	return &ConfigSnapshots{
		self: cfg,
	}
}

type ConfigSnapshots struct {
	self *config.Snapshots
}

func (e *ConfigSnapshots) SetAutoRecover(val bool) error {
	e.self.AutoRecover = val
	return config.Save(e.self)
}

func (e *ConfigSnapshots) SetCreateWhenSyncing(val bool) error {
	e.self.CreateWhenSyncing = val
	return config.Save(e.self)
}

func (e *ConfigSnapshots) SetMaxAuto(val int) error {
	if val < 2 {
		return errors.New("min max auto: 2")
	}
	e.self.MaxAuto = val
	return config.Save(e.self)
}

func (e *ConfigSnapshots) AutoRecover() bool {
	return e.self.AutoRecover
}

func (e *ConfigSnapshots) CreateWhenSyncing() bool {
	return e.self.CreateWhenSyncing
}

func (e *ConfigSnapshots) MaxAuto() int {
	return e.self.MaxAuto
}

func Snapshots(remoteName, filterAuto string) (*SnapshotSlice, error) {
	var remName shared.RemoteName
	remName.FromString(remoteName)
	var filterAutoConv snapshot.SnapshotsFilterAuto
	filterAutoConv.FromString(filterAuto)
	snaps, err := repository.Snap.Snapshots(remName, filterAutoConv)
	return newSnapshotSlice(snaps), err
}

func SnapshotByID(id string) (*Snapshot, error) {
	snap, err := repository.Snap.Snapshot(id)
	if shared.IsNil(snap) {
		return nil, shared.NewErrSnapshotNotFound(_packageName, id)
	}
	return newSnapshot(snap), err
}

func CreateSnapshot(
	alias string,
	accountID string,
	likedAlbums, likedArtists, likedTracks, playlists bool) (*Snapshot, error) {

	var (
		snap snapshot.Snapshot
		err  error
	)

	if err = execTask(0, func(ctx context.Context) error {
		account, err := accountByID(accountID)
		if err != nil {
			return err
		}

		actions, err := account.Actions()
		if err != nil {
			return err
		}

		snap, err = repository.Snap.Create(account.RemoteName(), alias, false)
		if err != nil {
			return err
		}

		if likedAlbums {
			action := actions.LikedAlbums()
			entities, err := action.Liked(ctx)
			if err != nil {
				return err
			}
			var conv shared.RemoteIDSlice[shared.RemoteEntity]
			conv.FromMap(entities)
			if err := snap.SetLikedAlbums(conv); err != nil {
				return err
			}
		}

		if likedArtists {
			action := actions.LikedArtists()
			entities, err := action.Liked(ctx)
			if err != nil {
				return err
			}
			var conv shared.RemoteIDSlice[shared.RemoteEntity]
			conv.FromMap(entities)
			if err := snap.SetLikedArtists(conv); err != nil {
				return err
			}
		}

		if likedTracks {
			action := actions.LikedTracks()
			entities, err := action.Liked(ctx)
			if err != nil {
				return err
			}
			var conv shared.RemoteIDSlice[shared.RemoteEntity]
			conv.FromMap(entities)
			if err := snap.SetLikedTracks(conv); err != nil {
				return err
			}
		}

		if playlists {
			action := actions.Playlist()
			playlists, err := action.MyPlaylists(ctx)
			if err != nil {
				return err
			}
			for _, pl := range playlists {
				tracks, err := pl.Tracks(ctx)
				if err != nil {
					return err
				}
				var conv shared.RemoteIDSlice[shared.RemoteTrack]
				conv.FromMap(tracks)
				vis, _ := pl.IsVisible()
				if _, err = snap.AddPlaylist(pl.Name(), pl.Description(), vis, conv); err != nil {
					return err
				}
			}
		}

		return err
	}); err != nil {
		return nil, err
	}

	return newSnapshot(snap), err
}

// Transfer snapshot to any account. Returns new snapshot ID.
func CrossShot(snapshotId, targetRemoteName string) (string, error) {
	var (
		snapshotID string
	)
	err := execTask(0, func(ctx context.Context) error {
		// Get snapshot.
		snap, err := SnapshotByID(snapshotId)
		if err != nil {
			return err
		}
		snapshot := snap.self

		// Get snapshot origin.
		origin, ok := _remotes[snapshot.RemoteName()]
		if !ok {
			return shared.NewErrRemoteNotFound(_packageName, snapshot.RemoteName())
		}
		originActions, err := origin.Actions()
		if err != nil {
			return err
		}

		// Get target.
		var targetRemoteNameConv shared.RemoteName
		targetRemoteNameConv.FromString(targetRemoteName)

		// Same remotes.
		if origin.Name() == targetRemoteNameConv {
			return nil
		}

		target, ok := _remotes[targetRemoteNameConv]
		if !ok {
			return shared.NewErrRemoteNotFound(_packageName, targetRemoteNameConv)
		}
		targetSnap, err := repository.Snap.Create(target.Name(), snap.Alias(), false)
		if err != nil {
			return err
		}
		snapshotID = targetSnap.ID()
		defer func() {
			if err != nil {
				_ = targetSnap.Delete()
			}
		}()

		if snap.LikedAlbumsRestoreable() {
			lnk, err := linkerimpl.NewAlbums()
			if err != nil {
				return err
			}
			originIds, err := snapshot.LikedAlbums()
			if err != nil {
				return err
			}
			getWrap := func(ctx context.Context, id shared.RemoteID) (shared.RemoteEntity, error) {
				return originActions.Album(ctx, id)
			}
			converted, err := transferIds(ctx, lnk, getWrap, originIds, target.Name())
			if err != nil {
				return err
			}
			if err := targetSnap.SetLikedAlbums(converted); err != nil {
				return err
			}
		}

		if snap.LikedArtistsRestoreable() {
			lnk, err := linkerimpl.NewArtists()
			if err != nil {
				return err
			}
			originIds, err := snapshot.LikedArtists()
			if err != nil {
				return err
			}
			getWrap := func(ctx context.Context, id shared.RemoteID) (shared.RemoteEntity, error) {
				return originActions.Artist(ctx, id)
			}
			converted, err := transferIds(ctx, lnk, getWrap, originIds, target.Name())
			if err != nil {
				return err
			}
			if err := targetSnap.SetLikedArtists(converted); err != nil {
				return err
			}
		}

		lnkTracks, err := linkerimpl.NewTracks()
		if err != nil {
			return err
		}
		getTrackWrap := func(ctx context.Context, id shared.RemoteID) (shared.RemoteEntity, error) {
			return originActions.Track(ctx, id)
		}
		if snap.LikedTracksRestoreable() {
			originIds, err := snapshot.LikedTracks()
			if err != nil {
				return err
			}
			converted, err := transferIds(ctx, lnkTracks, getTrackWrap, originIds, target.Name())
			if err != nil {
				return err
			}
			if err := targetSnap.SetLikedTracks(converted); err != nil {
				return err
			}
		}

		if snap.PlaylistsRestoreable() {
			pls, err := snap.Playlists()
			if err != nil {
				return err
			}
			plsUnwrap := pls.wrap.items
			for _, plWrap := range plsUnwrap {
				plUnwrap := plWrap.self
				originIds, err := plUnwrap.Tracks()
				if err != nil {
					return err
				}
				converted, err := transferIds(ctx, lnkTracks, getTrackWrap, originIds, target.Name())
				if err != nil {
					return err
				}
				_, err = targetSnap.AddPlaylist(plUnwrap.Name(), plUnwrap.Description(), plUnwrap.IsVisible(), converted)
				if err != nil {
					return err
				}
			}
		}

		return err
	})

	return snapshotID, err
}

func newSnapshot(self snapshot.Snapshot) *Snapshot {
	return &Snapshot{
		self: self,
	}
}

type Snapshot struct {
	self snapshot.Snapshot
}

func (e Snapshot) ID() string {
	return e.self.ID()
}

func (e Snapshot) Auto() bool {
	return e.self.Auto()
}

func (e Snapshot) Alias() string {
	return e.self.Alias()
}

func (e Snapshot) SetAlias(val string) error {
	return e.self.SetAlias(val)
}

func (e Snapshot) LikedAlbumsRestoreable() bool {
	return e.self.LikedAlbumsRestoreable()
}

func (e Snapshot) LikedAlbumsCount() (int, error) {
	return e.self.LikedAlbumsCount()
}

func (e Snapshot) RestoreLikedAlbums(merge bool, accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		actions, err := accountActionsByID(accountID)
		if err != nil {
			return err
		}
		return e.self.RestoreLikedAlbums(ctx, merge, actions.LikedAlbums())
	})
}

func (e Snapshot) LikedArtistsRestoreable() bool {
	return e.self.LikedArtistsRestoreable()
}

func (e Snapshot) LikedArtistsCount() (int, error) {
	return e.self.LikedArtistsCount()
}

func (e Snapshot) RestoreLikedArtists(merge bool, accountID string) error {
	return execTask(9, func(ctx context.Context) error {
		actions, err := accountActionsByID(accountID)
		if err != nil {
			return err
		}
		return e.self.RestoreLikedArtists(ctx, merge, actions.LikedArtists())
	})
}

func (e Snapshot) LikedTracksRestoreable() bool {
	return e.self.LikedTracksRestoreable()
}

func (e Snapshot) LikedTracksCount() (int, error) {
	return e.self.LikedTracksCount()
}

func (e Snapshot) RestoreLikedTracks(merge bool, accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		actions, err := accountActionsByID(accountID)
		if err != nil {
			return err
		}
		return e.self.RestoreLikedTracks(ctx, merge, actions.LikedTracks())
	})
}

func (e Snapshot) PlaylistsRestoreable() bool {
	return e.self.PlaylistsRestoreable()
}

func (e Snapshot) PlaylistsCount() (int, error) {
	return e.self.PlaylistsCount()
}

func (e Snapshot) Playlists() (*SnapshotPlaylistSlice, error) {
	snaps, err := e.self.Playlists()
	return newSnapshotPlaylistSlice(snaps), err
}

func (e Snapshot) Playlist(id string) (*SnapshotPlaylist, error) {
	snap, err := e.self.Playlist(id)
	return newSnapshotPlaylist(snap), err
}

func (e Snapshot) RestorePlaylists(merge bool, accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		actions, err := accountActionsByID(accountID)
		if err != nil {
			return err
		}
		return e.self.RestorePlaylists(ctx, merge, actions.Playlist())
	})
}

func (e Snapshot) CreatedAt() string {
	return e.self.CreatedAt().String()
}

func (e *Snapshot) Delete() error {
	return execTask(0, func(ctx context.Context) error {
		err := e.self.Delete()
		if err == nil {
			e.self = nil
		}
		return err
	})
}

func newSnapshotPlaylist(self snapshot.Playlist) *SnapshotPlaylist {
	return &SnapshotPlaylist{
		self: self,
	}
}

type SnapshotPlaylist struct {
	self snapshot.Playlist
}

func (e SnapshotPlaylist) ID() string {
	return e.self.ID()
}

func (e SnapshotPlaylist) Name() string {
	return e.self.Name()
}

func (e SnapshotPlaylist) IsVisible() bool {
	return e.self.IsVisible()
}

func (e SnapshotPlaylist) Description() string {
	desc := e.self.Description()
	if desc != nil {
		return *desc
	}
	return ""
}

func (e SnapshotPlaylist) Restore(accountID string) error {
	return execTask(0, func(ctx context.Context) error {
		actions, err := accountActionsByID(accountID)
		if err != nil {
			return err
		}
		return e.self.Restore(ctx, actions.Playlist())
	})
}

func (e *SnapshotPlaylist) Delete() error {
	return execTask(0, func(ctx context.Context) error {
		err := e.self.Delete()
		if err == nil {
			e.self = nil
		}
		return err
	})
}

func newSnapshotSlice(snaps []snapshot.Snapshot) *SnapshotSlice {
	converted := make([]Snapshot, len(snaps))
	for i := range converted {
		converted[i] = *newSnapshot(snaps[i])
	}
	return &SnapshotSlice{
		wrap: &wrappedSlice[Snapshot]{
			items: converted,
		},
	}
}

type SnapshotSlice struct {
	wrap *wrappedSlice[Snapshot]
}

func (e SnapshotSlice) Item(i int) *Snapshot {
	return e.wrap.Item(i)
}
func (e SnapshotSlice) Len() int {
	return e.wrap.Len()
}

func newSnapshotPlaylistSlice(snaps []snapshot.Playlist) *SnapshotPlaylistSlice {
	converted := make([]SnapshotPlaylist, len(snaps))
	for i := range converted {
		converted[i] = *newSnapshotPlaylist(snaps[i])
	}
	return &SnapshotPlaylistSlice{
		wrap: &wrappedSlice[SnapshotPlaylist]{
			items: converted,
		},
	}
}

type SnapshotPlaylistSlice struct {
	wrap *wrappedSlice[SnapshotPlaylist]
}

func (e SnapshotPlaylistSlice) Item(i int) *SnapshotPlaylist {
	return e.wrap.Item(i)
}
func (e SnapshotPlaylistSlice) Len() int {
	return e.wrap.Len()
}
