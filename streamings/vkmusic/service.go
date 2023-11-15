package vkmusic

import (
	"context"

	"github.com/oklookat/govkm"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

var (
	_repo streaming.Database
)

const (
	ServiceName streaming.ServiceName = "VK Music"
)

type Service struct {
}

func (s *Service) Boot(repo streaming.Database) error {
	_repo = repo
	return nil
}

func (s Service) Name() streaming.ServiceName {
	return ServiceName
}

func (s Service) Database() streaming.Database {
	return _repo
}

func (s Service) AssignAccountActions(account streaming.Account) (streaming.AccountActions, error) {
	return newAccountActions(account)
}

func (s Service) Actions() (streaming.ServiceActions, error) {
	accounts, err := _repo.Accounts(context.Background())
	if err != nil || len(accounts) == 0 {
		return nil, err
	}

	var client *govkm.Client
	for i := range accounts {
		client, err = getClient(accounts[i])
		if err != nil {
			continue
		}
		break
	}

	if client == nil {
		return nil, shared.ErrNoRemoteActions
	}

	return newActions(client), nil
}
