package repository

import (
	"context"
	"strings"

	"github.com/oklookat/synchro/shared"
)

type Remote struct {
	HID      shared.RepositoryID `db:"id"`
	HName    shared.RemoteName   `db:"name"`
	HEnabled bool                `db:"is_enabled"`

	//
	parent shared.Remote `db:"-" json:"-"`
}

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

	const query2 = `INSERT INTO remote (id, name) VALUES (?, ?) RETURNING *;`
	remote, err = dbGetOne[Remote](context.Background(), query2, genRepositoryID(), rem.Name())
	if err != nil {
		return nil, err
	}

	remote.parent = rem
	return remote, err
}

func (e Remote) ID() shared.RepositoryID {
	return e.HID
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
	alias = strings.TrimSpace(alias)
	if len(alias) == 0 {
		alias = e.Name().String() + " " + shared.GenerateULID()
	}

	const query = "INSERT INTO account (id, remote_name, alias, auth, added_at) VALUES (?, ?, ?, ?, ?) RETURNING *"
	acc, err := dbGetOne[Account](context.Background(), query, genRepositoryID(), e.Name(), alias, auth, shared.TimestampNow())
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

func (e *Remote) Account(id shared.RepositoryID) (shared.Account, error) {
	return dbGetOne[Account](context.Background(), "SELECT * FROM account WHERE remote_name=? AND id=? LIMIT 1", e.Name(), id)
}

func (e Remote) Actions() (shared.RemoteActions, error) {
	return e.parent.Actions()
}
