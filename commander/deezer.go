package commander

import (
	"context"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/remote/deezer"
	"github.com/oklookat/synchro/shared"
)

func NewConfigDeezer() *ConfigDeezer {
	cfg := &config.Deezer{}
	if err := config.Get(cfg); err != nil {
		cfg.Default()
	}
	return &ConfigDeezer{
		self: cfg,
	}
}

type ConfigDeezer struct {
	self *config.Deezer
}

func (e *ConfigDeezer) SetHost(val string) error {
	e.self.Host = val
	return config.Save(e.self)
}

func (e *ConfigDeezer) SetPort(val int) error {
	if err := shared.IsValidPort(val); err != nil {
		return err
	}
	e.self.Port = val
	return config.Save(e.self)
}

func (e *ConfigDeezer) Host() string {
	return e.self.Host
}

func (e *ConfigDeezer) Port() int {
	return e.self.Port
}

func NewDeezer() *Deezer {
	return &Deezer{}
}

type Deezer struct {
}

func (e Deezer) RemoteName() string {
	return deezer.RemoteName.String()
}

func (e Deezer) NewAccount(alias, clientID, clientSecret string, deadlineSeconds int, onURL OnUrler) (string, error) {
	var (
		account shared.Account
		err     error
	)
	if err := execTask(deadlineSeconds, func(ctx context.Context) error {
		account, err = deezer.NewAccount(ctx, alias, clientID, clientSecret, onURL.OnURL)
		return err
	}); err != nil {
		return "", err
	}
	return account.ID().String(), err
}

func (e Deezer) Reauth(accountId, clientID string, clientSecret string, deadlineSeconds int, onURL OnUrler) error {
	return execTask(deadlineSeconds, func(ctx context.Context) error {
		accById, err := accountByID(accountId)
		if err != nil {
			return err
		}
		return deezer.Reauth(ctx, accById, clientID, clientSecret, onURL.OnURL)
	})
}
