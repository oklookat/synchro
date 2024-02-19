package commander

import (
	"context"

	"github.com/oklookat/synchro/remote/yandexmusic"
	"github.com/oklookat/synchro/shared"
)

func NewYandexMusic() *YandexMusic {
	return &YandexMusic{}
}

type YandexMusic struct {
}

func (e YandexMusic) RemoteName() string {
	return yandexmusic.RemoteName.String()
}

func (e YandexMusic) NewAccount(alias string, deadlineSeconds int, onUrlCode OnUrlCoder) (string, error) {
	var (
		account shared.Account
		err     error
	)
	if err := execTask(deadlineSeconds, func(ctx context.Context) error {
		account, err = yandexmusic.NewAccount(ctx, alias, func(url, code string) {
			onUrlCode.OnUrlCode(url, code)
		})
		return err
	}); err != nil {
		return "", err
	}
	return account.ID().String(), err
}

func (e YandexMusic) Reauth(accountId string, login string, deadlineSeconds int, onURL OnUrlCoder) error {
	return execTask(deadlineSeconds, func(ctx context.Context) error {
		accById, err := accountByID(accountId)
		if err != nil {
			return err
		}
		return yandexmusic.Reauth(ctx, accById, login, onURL.OnUrlCode)
	})
}
