package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

func AccountByID(id uint64) (streaming.Account, error) {
	const query = "SELECT * FROM account WHERE id = ? LIMIT 1"
	return dbGetOne[Account](context.Background(), query, id)
}

type Account struct {
	HID         uint64                `db:"id"`
	HRemoteName streaming.ServiceName `db:"service_name"`
	HAuth       string                `db:"auth"`
	HAlias      string                `db:"alias"`
	HAddedAt    int64                 `db:"added_at"`
}

func (e Account) ID() string {
	return strconv.FormatUint(e.HID, 10)
}

func (e Account) ServiceName() streaming.ServiceName {
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

func (e Account) AddedAt() time.Time {
	return shared.Time(e.HAddedAt)
}

func (e Account) Delete() error {
	const query = "DELETE FROM account WHERE id=?"
	_, err := dbExec(context.Background(), query, e.HID)
	return err
}

func (e Account) Database() (streaming.Database, error) {
	rem, ok := Remotes[e.HRemoteName]
	if !ok {
		return nil, errors.New("service not found")
	}
	return rem.Database(), nil
}

func (e *Account) Actions() (streaming.AccountActions, error) {
	rem, ok := Remotes[e.HRemoteName]
	if !ok {
		return nil, errors.New("service not found")
	}
	return rem.AssignAccountActions(e)
}
