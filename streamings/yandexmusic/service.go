package yandexmusic

import (
	"context"

	"github.com/oklookat/synchro/streaming"
)

var (
	_repo streaming.Database
)

const (
	ServiceName streaming.ServiceName = "Yandex.Music"
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

	client, err := getClient(accounts[0])
	if err != nil {
		return nil, err
	}

	return newActions(client), nil
}
