package repository

import (
	"context"
	"strconv"

	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/shared"
)

func RemoteByName(name shared.RemoteName) (shared.Remote, error) {
	parent, ok := Remotes[name]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(_packageName, name)
	}
	return parent, nil
}

func newOrExistingRemote(rem shared.Remote) (*Remote, error) {
	const query = "SELECT * FROM remote WHERE name=? LIMIT 1"
	remote, err := dbGetOne[Remote](context.Background(), query, rem.Name())
	if err != nil {
		return nil, err
	}
	if remote != nil {
		remote.parent = rem
		return remote, err
	}

	const query2 = `INSERT INTO remote (name) VALUES (?) RETURNING *;`
	remote, err = dbGetOne[Remote](context.Background(), query2, rem.Name())
	if err != nil {
		return nil, err
	}

	remote.parent = rem
	return remote, err
}

type Remote struct {
	HID      uint64            `db:"id"`
	HName    shared.RemoteName `db:"name"`
	HEnabled bool              `db:"is_enabled"`

	//
	parent shared.Remote `db:"-" json:"-"`
}

func (e Remote) ID() string {
	return strconv.FormatUint(e.HID, 10)
}

func (e Remote) Enabled() bool {
	return e.HEnabled
}

func (e *Remote) SetEnabled(val bool) error {
	const query = "UPDATE remote SET is_enabled=? WHERE name=?"
	_, err := dbExec(context.Background(), query, val, e.Name())
	if err == nil {
		e.HEnabled = val
	}
	return err
}

func (e Remote) Name() shared.RemoteName {
	return e.HName
}

func (e *Remote) CreateAccount(alias string, auth string) (shared.Account, error) {
	const query = "INSERT INTO account (remote_name, alias, auth, added_at) VALUES (?, ?, ?, ?) RETURNING *"
	acc, err := dbGetOne[Account](context.Background(), query, e.Name(), alias, auth, shared.TimestampNow())
	if err != nil {
		return nil, err
	}

	// Init settings.
	_, err = acc.Settings()
	if err != nil {
		return nil, err
	}

	return acc, err
}

func (e *Remote) Accounts(ctx context.Context) ([]shared.Account, error) {
	const query = "SELECT * FROM account WHERE remote_name=?"
	return dbGetManyConvert[Account, shared.Account](ctx, nil, query, e.Name())
}

func (e *Remote) Account(id string) (shared.Account, error) {
	return dbGetOne[Account](context.Background(), "SELECT * FROM account WHERE remote_name=? AND id=? LIMIT 1", e.Name(), id)
}

func (e Remote) Actions() (shared.RemoteActions, error) {
	log := _log.AddField("remote", e.Name().String())
	acts, err := e.parent.Actions()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return wrappedRemoteActions{
		log:  &log,
		acts: acts,
	}, err
}

type wrappedRemoteActions struct {
	log  *logger.Logger
	acts shared.RemoteActions
}

func (e wrappedRemoteActions) Album(ctx context.Context, id shared.RemoteID) (shared.RemoteAlbum, error) {
	ent, err := e.acts.Album(ctx, id)
	if err != nil {
		e.log.Error("album: " + err.Error())
	}
	return ent, err
}

func (e wrappedRemoteActions) Artist(ctx context.Context, id shared.RemoteID) (shared.RemoteArtist, error) {
	ent, err := e.acts.Artist(ctx, id)
	if err != nil {
		e.log.Error("artist: " + err.Error())
	}
	return ent, err
}

func (e wrappedRemoteActions) Track(ctx context.Context, id shared.RemoteID) (shared.RemoteTrack, error) {
	ent, err := e.acts.Track(ctx, id)
	if err != nil {
		e.log.Error("track: " + err.Error())
	}
	return ent, err
}

func (e wrappedRemoteActions) SearchAlbums(ctx context.Context, what shared.RemoteAlbum) ([10]shared.RemoteAlbum, error) {
	res, err := e.acts.SearchAlbums(ctx, what)
	if err != nil {
		e.log.Error("searchAlbums: " + err.Error())
	}
	return res, err
}

func (e wrappedRemoteActions) SearchArtists(ctx context.Context, what shared.RemoteArtist) ([10]shared.RemoteArtist, error) {
	res, err := e.acts.SearchArtists(ctx, what)
	if err != nil {
		e.log.Error("searchArtists: " + err.Error())
	}
	return res, err
}

func (e wrappedRemoteActions) SearchTracks(ctx context.Context, what shared.RemoteTrack) ([10]shared.RemoteTrack, error) {
	res, err := e.acts.SearchTracks(ctx, what)
	if err != nil {
		e.log.Error("searchTracks: " + err.Error())
	}
	return res, err
}
