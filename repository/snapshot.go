package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/snapshot"
)

var Snap Snapshotter

type Snapshotter struct {
}

func (e Snapshotter) Create(in shared.RemoteName, alias string, auto bool) (snapshot.Snapshot, error) {
	cfg := &config.Snapshots{}
	if err := config.Get(cfg); err != nil {
		cfg.Default()
	}

	const query = "INSERT INTO snapshot (remote_name, alias, auto, created_at) VALUES (?, ?, ?, ?) RETURNING *;"
	return dbGetOne[Snapshot](context.Background(), query, in, alias, auto, shared.TimestampNow())
}

func (e Snapshotter) DeleteOldestAuto(in shared.RemoteName, max int) error {
	query := `
	DELETE FROM snapshot
	WHERE remote_name = ?
	AND (
		SELECT COUNT(*)
		FROM snapshot
		WHERE remote_name = ?
	) > %d
	AND id IN (
		SELECT id
		FROM snapshot
		WHERE remote_name = ?
		ORDER BY created_at ASC
		LIMIT (
			SELECT COUNT(*) - %d
			FROM snapshot
			WHERE remote_name = ?
		)
	);	
`
	query = fmt.Sprintf(query, max, max)
	_, err := dbExec(context.Background(), query, in, in, in, in)
	return err
}

func (e Snapshotter) Snapshots(
	remote shared.RemoteName,
	filter snapshot.SnapshotsFilterAuto,
) ([]snapshot.Snapshot, error) {
	var args []any
	query := "SELECT * FROM snapshot"
	if len(remote) > 0 {
		query += " WHERE remote_name = ?"
		args = append(args, remote.String())
	}
	if filter == snapshot.SnapshotsFilterAutoAuto {
		query += " AND auto = 1"
	} else if filter == snapshot.SnapshotsFilterAutoManual {
		query += " AND auto = 0"
	}
	query += " ORDER BY created_at DESC"

	return dbGetManyConvert[Snapshot, snapshot.Snapshot](context.Background(), nil, query, args...)
}

func (e Snapshotter) Snapshot(id string) (snapshot.Snapshot, error) {
	uid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	const query = "SELECT * FROM snapshot WHERE id = ? LIMIT 1"
	return dbGetOne[Snapshot](context.Background(), query, uid)
}

type Snapshot struct {
	HID                      uint64            `db:"id"`
	HRemoteName              shared.RemoteName `db:"remote_name"`
	HAlias                   string            `db:"alias"`
	HAuto                    bool              `db:"auto"`
	HRestoreableLikedAlbums  bool              `db:"restoreable_liked_albums"`
	HRestoreableLikedArtists bool              `db:"restoreable_liked_artists"`
	HRestoreableLikedTracks  bool              `db:"restoreable_liked_tracks"`
	HRestoreablePlaylists    bool              `db:"restoreable_playlists"`
	HCreatedAt               int64             `db:"created_at"`
}

func (e Snapshot) ID() string {
	return strconv.FormatUint(e.HID, 10)
}

func (e Snapshot) RemoteName() shared.RemoteName {
	return e.HRemoteName
}

func (e Snapshot) Auto() bool {
	return e.HAuto
}

func (e Snapshot) Alias() string {
	return e.HAlias
}

func (e Snapshot) SetAlias(val string) error {
	const query = "UPDATE snapshot SET alias = ? WHERE id = ?"
	_, err := dbExec(context.Background(), query, strings.TrimSpace(val), e.HID)
	return err
}

func (e Snapshot) LikedAlbumsRestoreable() bool {
	return e.HRestoreableLikedAlbums
}

func (e Snapshot) LikedAlbumsCount() (int, error) {
	return execSnapshotGetCountQuery("snapshot_liked_album", e.HID)
}

func (e Snapshot) LikedAlbums() ([]shared.RemoteID, error) {
	res, err := getSnapshotLikes(context.Background(), "snapshot_liked_album", e.HID)
	if err != nil {
		return nil, err
	}
	return res.Ids(), err
}

func (e *Snapshot) SetLikedAlbums(ids []shared.RemoteID) error {
	if err := snapshotExecSetIds("snapshot_liked_album", e.HID, ids); err != nil {
		return err
	}
	if err := snapshotExecSetRestoreable("snapshot", "restoreable_liked_albums", e.HID); err != nil {
		return err
	}
	e.HRestoreableLikedAlbums = true
	return nil
}

func (e Snapshot) RestoreLikedAlbums(ctx context.Context, merge bool, action shared.LikedActions) error {
	ids, err := e.LikedAlbums()
	if err != nil {
		return err
	}
	return snapshotGetLikedRestorer(e.LikedAlbumsRestoreable(), ids)(ctx, merge, action)
}

func (e Snapshot) LikedArtistsRestoreable() bool {
	return e.HRestoreableLikedArtists
}

func (e Snapshot) LikedArtistsCount() (int, error) {
	return execSnapshotGetCountQuery("snapshot_liked_artist", e.HID)
}

func (e Snapshot) LikedArtists() ([]shared.RemoteID, error) {
	res, err := getSnapshotLikes(context.Background(), "snapshot_liked_artist", e.HID)
	if err != nil {
		return nil, err
	}
	return res.Ids(), err
}

func (e *Snapshot) SetLikedArtists(ids []shared.RemoteID) error {
	if err := snapshotExecSetIds("snapshot_liked_artist", e.HID, ids); err != nil {
		return err
	}
	if err := snapshotExecSetRestoreable("snapshot", "restoreable_liked_artists", e.HID); err != nil {
		return err
	}
	e.HRestoreableLikedArtists = true
	return nil
}

func (e Snapshot) RestoreLikedArtists(ctx context.Context, merge bool, action shared.LikedActions) error {
	ids, err := e.LikedArtists()
	if err != nil {
		return err
	}
	return snapshotGetLikedRestorer(e.LikedArtistsRestoreable(), ids)(ctx, merge, action)
}

func (e Snapshot) LikedTracksRestoreable() bool {
	return e.HRestoreableLikedTracks
}

func (e Snapshot) LikedTracksCount() (int, error) {
	return execSnapshotGetCountQuery("snapshot_liked_track", e.HID)
}

func (e Snapshot) LikedTracks() ([]shared.RemoteID, error) {
	res, err := getSnapshotLikes(context.Background(), "snapshot_liked_track", e.HID)
	if err != nil {
		return nil, err
	}
	return res.Ids(), err
}

func (e *Snapshot) SetLikedTracks(ids []shared.RemoteID) error {
	if err := snapshotExecSetIds("snapshot_liked_track", e.HID, ids); err != nil {
		return err
	}
	if err := snapshotExecSetRestoreable("snapshot", "restoreable_liked_tracks", e.HID); err != nil {
		return err
	}
	e.HRestoreableLikedTracks = true
	return nil
}

func (e Snapshot) RestoreLikedTracks(ctx context.Context, merge bool, action shared.LikedActions) error {
	ids, err := e.LikedTracks()
	if err != nil {
		return err
	}
	return snapshotGetLikedRestorer(e.LikedTracksRestoreable(), ids)(ctx, merge, action)
}

func (e Snapshot) PlaylistsRestoreable() bool {
	return e.HRestoreablePlaylists
}

func (e Snapshot) PlaylistsCount() (int, error) {
	return execSnapshotGetCountQuery("snapshot_playlist", e.HID)
}

func (e Snapshot) Playlist(id string) (snapshot.Playlist, error) {
	const query = "SELECT * FROM snapshot_playlist WHERE snapshot_id = ? AND id = ? LIMIT 1"
	return dbGetOne[SnapshotPlaylist](context.Background(), query, e.HID, id)
}

func (e Snapshot) Playlists() ([]snapshot.Playlist, error) {
	const query = "SELECT * FROM snapshot_playlist WHERE snapshot_id = ? ORDER BY created_at DESC"
	return dbGetManyConvert[SnapshotPlaylist, snapshot.Playlist](context.Background(), nil, query, e.HID)
}

func (e *Snapshot) AddPlaylist(name string,
	description *string,
	isVisible bool,
	tracks []shared.RemoteID) (snapshot.Playlist, error) {
	const query = "INSERT INTO snapshot_playlist (snapshot_id, name, is_visible, description, created_at) VALUES (?, ?, ?, ?, ?) RETURNING *"
	snap, err := dbGetOne[SnapshotPlaylist](context.Background(), query, e.HID, name, isVisible, description, shared.TimestampNow())
	if err != nil {
		return nil, err
	}
	if err := snap.SetTracks(tracks); err != nil {
		return nil, err
	}
	if err := snapshotExecSetRestoreable("snapshot", "restoreable_playlists", e.HID); err != nil {
		return nil, err
	}
	e.HRestoreablePlaylists = true
	return snap, err
}

func (e Snapshot) RestorePlaylists(ctx context.Context, merge bool, action shared.PlaylistActions) error {
	if !merge {
		pls, err := action.MyPlaylists(ctx)
		if err != nil {
			return err
		}
		if len(pls) > 0 {
			var conv shared.RemoteIDSlice[shared.RemotePlaylist]
			conv.FromMap(pls)
			if err = action.Delete(ctx, conv); err != nil {
				return err
			}
		}
	}
	pls, err := e.Playlists()
	if err != nil {
		return err
	}
	for _, pl := range pls {
		if err = pl.Restore(ctx, action); err != nil {
			return err
		}
	}
	return err
}

func (e Snapshot) CreatedAt() time.Time {
	return shared.Time(e.HCreatedAt)
}

func (e Snapshot) Delete() error {
	_, err := dbExec(context.Background(), "DELETE FROM snapshot WHERE id = ?", e.HID)
	return err
}

type SnapshotPlaylist struct {
	HID          uint64  `db:"id"`
	HSnapshotID  uint64  `db:"snapshot_id"`
	HName        string  `db:"name"`
	HIsVisible   bool    `db:"is_visible"`
	HDescription *string `db:"description"`
	HCreatedAt   int64   `db:"created_at"`
}

func (e SnapshotPlaylist) ID() string {
	return strconv.FormatUint(e.HID, 10)
}

func (e SnapshotPlaylist) Name() string {
	return e.HName
}

func (e SnapshotPlaylist) IsVisible() bool {
	return e.HIsVisible
}

func (e SnapshotPlaylist) Description() *string {
	return e.HDescription
}

func (e SnapshotPlaylist) Tracks() ([]shared.RemoteID, error) {
	res, err := getSnapshotLikes(context.Background(), "snapshot_playlist_track", e.HID)
	if err != nil {
		return nil, err
	}
	return res.Ids(), err
}

func (e SnapshotPlaylist) SetTracks(ids []shared.RemoteID) error {
	return snapshotExecSetIds("snapshot_playlist_track", e.HID, ids)
}

func (e SnapshotPlaylist) Restore(ctx context.Context, action shared.PlaylistActions) error {
	ids, err := e.Tracks()
	if err != nil {
		return err
	}
	pl, err := action.Create(ctx, e.HName, e.HIsVisible, e.HDescription)
	if err != nil {
		return err
	}
	return pl.AddTracks(ctx, ids)
}

func (e SnapshotPlaylist) Delete() error {
	const query = "DELETE FROM snapshot_playlist WHERE id = ?"
	_, err := dbExec(context.Background(), query, e.HID)
	return err
}

func (e SnapshotPlaylist) CreatedAt() time.Time {
	return shared.Time(e.HCreatedAt)
}

func getSnapshotLikes(ctx context.Context, tableName string, snapshotID uint64) (snapshotLikedSlice, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE snapshot_id = ?", tableName)
	return dbGetMany[SnapshotLiked](ctx, query, nil, snapshotID)
}

type SnapshotLiked struct {
	HID         uint64          `db:"id"`
	HSnapshotID uint64          `db:"snapshot_id"`
	HIdOnRemote shared.RemoteID `db:"id_on_remote"`
}

type snapshotLikedSlice []*SnapshotLiked

func (e snapshotLikedSlice) Ids() []shared.RemoteID {
	var ids []shared.RemoteID
	for i := range e {
		if e[i] == nil {
			continue
		}
		ids = append(ids, e[i].HIdOnRemote)
	}
	return ids
}

func snapshotExecSetIds(tableName string, snapID uint64, ids []shared.RemoteID) error {
	if len(ids) == 0 {
		return nil
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE snapshot_id = ?", tableName)
	_, err := dbExec(context.Background(), query, snapID)
	if err != nil {
		return err
	}

	for _, id := range ids {
		query = fmt.Sprintf("INSERT INTO %s (snapshot_id, id_on_remote) VALUES (?, ?)", tableName)
		_, err := dbExec(context.Background(), query, snapID, id)
		if err != nil {
			return err
		}
	}

	return err
}

func snapshotExecSetRestoreable(tableName, restoreableColumnName string, snapID uint64) error {
	query := fmt.Sprintf("UPDATE %s SET %s = 1 WHERE id = %d", tableName, restoreableColumnName, snapID)
	_, err := dbExec(context.Background(), query)
	return err
}

func snapshotGetLikedRestorer(restoreable bool, ids []shared.RemoteID) func(ctx context.Context, merge bool, action shared.LikedActions) error {
	return func(ctx context.Context, merge bool, action shared.LikedActions) error {
		if !restoreable {
			return nil
		}
		if shared.IsNil(action) {
			return errors.New("nil action")
		}
		if !merge {
			likes, err := action.Liked(ctx)
			if err != nil {
				return err
			}
			if len(likes) > 0 {
				var conv shared.RemoteIDSlice[shared.RemoteEntity]
				conv.FromMap(likes)
				if err = action.Unlike(ctx, conv); err != nil {
					return err
				}
			}
		}
		return action.Like(ctx, ids)
	}
}
