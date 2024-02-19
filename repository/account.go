package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/shared"
)

func AccountByID(id shared.RepositoryID) (shared.Account, error) {
	const query = "SELECT * FROM account WHERE id = ? LIMIT 1"
	return dbGetOne[Account](context.Background(), query, id)
}

type Account struct {
	HID         shared.RepositoryID `db:"id"`
	HRemoteName shared.RemoteName   `db:"remote_name"`
	HAuth       string              `db:"auth"`
	HAlias      string              `db:"alias"`
	HAddedAt    int64               `db:"added_at"`
	//
	theSettings *AccountSettings `db:"-" json:"-"`
}

func (e Account) ID() shared.RepositoryID {
	return e.HID
}

func (e Account) RemoteName() shared.RemoteName {
	return e.HRemoteName
}

func (e Account) Alias() string {
	return e.HAlias
}

func (e *Account) SetAlias(alias string) error {
	const query = "UPDATE account SET alias=? WHERE id=?"
	_, err := dbExec(context.Background(), query, alias, e.HID)
	if err == nil {
		e.HAlias = alias
	}
	return err
}

func (e Account) Auth() string {
	return e.HAuth
}

func (e *Account) SetAuth(auth string) error {
	const query = "UPDATE account SET auth=? WHERE id=?"
	_, err := dbExec(context.Background(), query, auth, e.HID)
	if err == nil {
		e.HAuth = auth
	}
	return err
}

func (e *Account) Settings() (shared.AccountSettings, error) {
	if e.theSettings == nil {
		setts, err := newAccountSettings(e.HID)
		if err != nil {
			return nil, err
		}
		e.theSettings = setts
	}
	return e.theSettings, nil
}

func (e Account) AddedAt() time.Time {
	return shared.Time(e.HAddedAt)
}

func (e Account) Delete() error {
	const query = "DELETE FROM account WHERE id=?"
	_, err := dbExec(context.Background(), query, e.HID)
	return err
}

func (e Account) Repository() (shared.RemoteRepository, error) {
	rem, ok := Remotes[e.HRemoteName]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(_packageName, e.HRemoteName)
	}
	return rem.Repository(), nil
}

func (e *Account) Actions() (shared.AccountActions, error) {
	rem, ok := Remotes[e.HRemoteName]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(_packageName, e.HRemoteName)
	}

	theLog := _log.
		AddField("accountID", e.ID()).
		AddField("alias", e.Alias()).
		AddField("remote", e.RemoteName().String())

	acts, err := rem.AssignAccountActions(e)
	if err != nil {
		theLog.Error("assignAccountActions: " + err.Error())
		return nil, err
	}

	return wrappedAccountActions{
		log:  &theLog,
		acts: acts,
	}, err
}

type wrappedAccountActions struct {
	log  *logger.Logger
	acts shared.AccountActions
}

func (e wrappedAccountActions) LikedAlbums() shared.LikedActions {
	log := e.log.AddField("actions", "LikedAlbums")
	return wrappedLikedActions{
		log:  &log,
		acts: e.acts.LikedAlbums(),
	}
}

func (e wrappedAccountActions) LikedArtists() shared.LikedActions {
	log := e.log.AddField("actions", "LikedArtists")
	return wrappedLikedActions{
		log:  &log,
		acts: e.acts.LikedArtists(),
	}
}

func (e wrappedAccountActions) LikedTracks() shared.LikedActions {
	log := e.log.AddField("actions", "LikedTracks")
	return wrappedLikedActions{
		log:  &log,
		acts: e.acts.LikedTracks(),
	}
}

func (e wrappedAccountActions) Playlist() shared.PlaylistActions {
	log := e.log.AddField("actions", "Playlist")
	return wrappedPlaylistActions{
		log:  &log,
		acts: e.acts.Playlist(),
	}
}

type wrappedLikedActions struct {
	log  *logger.Logger
	acts shared.LikedActions
}

func (e wrappedLikedActions) Liked(ctx context.Context) (map[shared.RemoteID]shared.RemoteEntity, error) {
	e.log.Info("Getting likes...")

	res, err := e.acts.Liked(ctx)
	if err != nil {
		e.log.Error("liked: " + err.Error())
		return res, err
	}
	filtRes := map[shared.RemoteID]shared.RemoteEntity{}
	for id := range res {
		if shared.IsNil(res[id]) {
			e.log.Warn("liked: nil RemoteEntity")
			continue
		}
		filtRes[id] = res[id]
	}
	return filtRes, err
}

func (e wrappedLikedActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	e.log.Info(fmt.Sprintf("Adding %d things...", len(ids)))

	fids := e.filterIds("like", ids)
	err := e.acts.Like(ctx, fids)
	if err != nil {
		e.log.Error("like: " + err.Error())
	}
	return err
}

func (e wrappedLikedActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	e.log.Info(fmt.Sprintf("Removing %d things...", len(ids)))

	fids := e.filterIds("unlike", ids)
	err := e.acts.Unlike(ctx, fids)
	if err != nil {
		e.log.Error("unlike: " + err.Error())
	}
	return err
}

func (e wrappedLikedActions) filterIds(issuer string, ids []shared.RemoteID) []shared.RemoteID {
	res := make([]shared.RemoteID, 0, len(ids))
	for _, id := range ids {
		if len(id) == 0 {
			e.log.Warn(fmt.Sprintf("filterIds (%s): empty id", issuer))
			continue
		}
		res = append(res, id)
	}
	return res
}

type wrappedPlaylistActions struct {
	log  *logger.Logger
	acts shared.PlaylistActions
}

func (e wrappedPlaylistActions) MyPlaylists(ctx context.Context) (map[shared.RemoteID]shared.RemotePlaylist, error) {
	res, err := e.acts.MyPlaylists(ctx)
	if err != nil {
		e.log.Error("myPlaylists: " + err.Error())
		return res, err
	}
	filtRes := map[shared.RemoteID]shared.RemotePlaylist{}
	for id := range res {
		if shared.IsNil(res[id]) {
			e.log.AddField("id", id.String()).Error("myPlaylists: nil playlist")
			continue
		}
		filtRes[id] = res[id]
	}
	return filtRes, err
}

func (e wrappedPlaylistActions) Create(
	ctx context.Context,
	name string,
	isVisible bool,
	description *string,
) (shared.RemotePlaylist, error) {
	if len(name) == 0 {
		const msg = "empty playlist name"
		e.log.Error(msg)
		return nil, errors.New(msg)
	}
	res, err := e.acts.Create(ctx, name, isVisible, description)
	if err != nil {
		e.log.Error("create: " + err.Error())
		return res, err
	}
	if shared.IsNil(res) {
		const msg = "nil playlist returned"
		e.log.Error(msg)
		return res, errors.New(msg)
	}
	return res, err
}

func (e wrappedPlaylistActions) Delete(ctx context.Context, ids []shared.RemoteID) error {
	e.log.Info(fmt.Sprintf("Removing %d playlists...", len(ids)))

	fids := e.filterIds("delete", ids)
	if err := e.acts.Delete(ctx, fids); err != nil {
		e.log.Error("delete: " + err.Error())
	}
	return nil
}

func (e wrappedPlaylistActions) Playlist(ctx context.Context, id shared.RemoteID) (shared.RemotePlaylist, error) {
	if len(id) == 0 {
		const msg = "get playlist: empty id"
		e.log.Error(msg)
		return nil, errors.New(msg)
	}
	res, err := e.acts.Playlist(ctx, id)
	if err != nil {
		e.log.Error("get playlist: " + err.Error())
		return res, err
	}
	if shared.IsNil(res) {
		e.log.Error("get playlist: nil playlist returned")
		return res, err
	}
	return res, err
}

func (e wrappedPlaylistActions) filterIds(issuer string, ids []shared.RemoteID) []shared.RemoteID {
	res := make([]shared.RemoteID, 0, len(ids))
	for _, id := range ids {
		if len(id) == 0 {
			e.log.Warn(fmt.Sprintf("filterIds (%s): empty id", issuer))
			continue
		}
		res = append(res, id)
	}
	return res
}
