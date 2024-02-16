package commander

import (
	"strconv"

	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
)

type wrappedSlice[T any] struct {
	items []T
}

func (e *wrappedSlice[_]) Len() int {
	return len(e.items)
}

func (e *wrappedSlice[T]) Item(i int) *T {
	if i >= 0 && i < len(e.items) {
		return &e.items[i]
	}
	return nil
}

func stringToRemoteIDPtr(val string) *shared.RemoteID {
	var remID *shared.RemoteID
	if len(val) > 0 {
		remID = (*shared.RemoteID)(&val)
	}
	return remID
}

func accountByID(id string) (shared.Account, error) {
	accID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}
	account, err := repository.AccountByID(accID)
	if err != nil {
		return nil, err
	}
	if shared.IsNil(account) {
		return nil, shared.NewErrAccountNotExists(_packageName, id)
	}
	return account, err
}

func accountActionsByID(accountID string) (shared.AccountActions, error) {
	account, err := accountByID(accountID)
	if err != nil {
		return nil, err
	}
	actions, err := account.Actions()
	return actions, err
}
