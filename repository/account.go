package repository

import (
	"context"
	"time"

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
		return nil, shared.NewErrRemoteNotFound(e.HRemoteName)
	}
	return rem.Repository(), nil
}

func (e *Account) Actions() (shared.AccountActions, error) {
	rem, ok := Remotes[e.HRemoteName]
	if !ok {
		return nil, shared.NewErrRemoteNotFound(e.HRemoteName)
	}

	return rem.AssignAccountActions(e)
}
