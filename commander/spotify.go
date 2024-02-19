package commander

import (
	"context"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/remote/spotify"
	"github.com/oklookat/synchro/shared"
)

func NewConfigSpotify() *ConfigSpotify {
	cfg := &config.Spotify{}
	if err := config.Get(cfg); err != nil {
		cfg.Default()
	}
	return &ConfigSpotify{
		self: cfg,
	}
}

type ConfigSpotify struct {
	self *config.Spotify
}

func (e *ConfigSpotify) SetHost(val string) error {
	e.self.Host = val
	return config.Save(e.self)
}

func (e *ConfigSpotify) SetPort(val int) error {
	if err := shared.IsValidPort(val); err != nil {
		return err
	}
	e.self.Port = val
	return config.Save(e.self)
}

func (e *ConfigSpotify) Host() string {
	return e.self.Host
}

func (e *ConfigSpotify) Port() int {
	return e.self.Port
}

func NewSpotify() *Spotify {
	return &Spotify{}
}

type Spotify struct {
}

func (e Spotify) RemoteName() string {
	return spotify.RemoteName.String()
}

func (e Spotify) NewAccount(alias, clientID, clientSecret string, deadlineSeconds int, onURL OnUrler) (string, error) {
	var (
		account shared.Account
		err     error
	)
	if err := execTask(deadlineSeconds, func(ctx context.Context) error {
		account, err = spotify.NewAccount(ctx, alias, clientID, clientSecret, onURL.OnURL)
		return err
	}); err != nil {
		return "", err
	}
	return account.ID().String(), nil
}

func (e Spotify) Reauth(accountId, clientID string, clientSecret string, deadlineSeconds int, onURL OnUrler) error {
	return execTask(deadlineSeconds, func(ctx context.Context) error {
		accById, err := accountByID(accountId)
		if err != nil {
			return err
		}
		return spotify.Reauth(ctx, accById, clientID, clientSecret, onURL.OnURL)
	})
}
