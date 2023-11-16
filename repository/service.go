package repository

import (
	"context"
	"strconv"

	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

func ServiceByName(name streaming.ServiceName) (streaming.Service, error) {
	parent, ok := Services[name]
	if !ok {
		return nil, shared.NewErrServiceNotFound(name)
	}
	return parent, nil
}

func newOrExistingServiceDatabase(rem streaming.Service) (*Service, error) {
	const query = "SELECT * FROM service WHERE name=? LIMIT 1"
	service, err := dbGetOne[Service](context.Background(), query, rem.Name())
	if err != nil {
		return nil, err
	}
	if service != nil {
		service.parent = rem
		return service, err
	}

	const query2 = `INSERT INTO service (name) VALUES (?) RETURNING *;`
	service, err = dbGetOne[Service](context.Background(), query2, rem.Name())
	if err != nil {
		return nil, err
	}

	service.parent = rem
	return service, err
}

type Service struct {
	HID      uint64                `db:"id"`
	HName    streaming.ServiceName `db:"name"`
	HEnabled bool                  `db:"is_enabled"`

	//
	parent streaming.Service `db:"-" json:"-"`
}

func (e Service) ID() string {
	return strconv.FormatUint(e.HID, 10)
}

func (e Service) Name() streaming.ServiceName {
	return e.HName
}

func (e *Service) CreateAccount(alias string, auth string) (streaming.Account, error) {
	if len(alias) == 0 {
		alias = shared.GenerateWord(8)
	}
	const query = "INSERT INTO account (service_name, alias, auth, added_at) VALUES (?, ?, ?, ?) RETURNING *"
	return dbGetOne[Account](context.Background(), query, e.Name(), alias, auth, shared.TimestampNow())
}

func (e *Service) Accounts(ctx context.Context) ([]streaming.Account, error) {
	const query = "SELECT * FROM account WHERE service_name=?"
	return dbGetManyConvert[Account, streaming.Account](ctx, nil, query, e.Name())
}

func (e *Service) Account(id string) (streaming.Account, error) {
	return dbGetOne[Account](context.Background(), "SELECT * FROM account WHERE service_name=? AND id=? LIMIT 1", e.Name(), id)
}

func (e Service) Actions() (streaming.ServiceActions, error) {
	return e.parent.Actions()
}
