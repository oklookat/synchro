package yandexmusic

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/oklookat/goym"
	"github.com/oklookat/synchro/shared"
)

var (
	_repo shared.RemoteRepository
)

const (
	RemoteName shared.RemoteName = "Yandex.Music"
)

type Remote struct {
}

func (s *Remote) Boot(repo shared.RemoteRepository) error {
	_repo = repo
	return nil
}

func (s Remote) Name() shared.RemoteName {
	return RemoteName
}

func (s Remote) Repository() shared.RemoteRepository {
	return _repo
}

func (s Remote) AssignAccountActions(account shared.Account) (shared.AccountActions, error) {
	return newAccountActions(account)
}

func (s Remote) Actions() (shared.RemoteActions, error) {
	accounts, err := _repo.Accounts(context.Background())
	if err != nil || len(accounts) == 0 {
		return nil, err
	}

	var client *goym.Client
	for i := range accounts {
		client, err = getClient(accounts[i])
		if err != nil {
			// Account not available.
			slog.Error("getClient: " + err.Error())
			continue
		}
		break
	}

	if client == nil {
		return nil, shared.ErrNoRemoteActions
	}

	return newActions(client), nil
}

func (e Remote) EntityURL(etype shared.EntityType, id shared.RemoteID) url.URL {
	return shared.GetEntityURL("https://music.yandex.ru", etype, id)
}
