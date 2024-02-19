package commander

import (
	"context"

	"github.com/oklookat/synchro/remote/zvuk"
	"github.com/oklookat/synchro/shared"
)

func NewZvuk() *Zvuk {
	return &Zvuk{}
}

type Zvuk struct {
}

func (e Zvuk) RemoteName() string {
	return zvuk.RemoteName.String()
}

func (e Zvuk) NewAccount(alias, token string) (string, error) {
	var (
		account shared.Account
		err     error
	)
	if err := execTask(0, func(ctx context.Context) error {
		account, err = zvuk.NewAccount(context.Background(), alias, token)
		return err
	}); err != nil {
		return "", err
	}
	return account.ID().String(), err
}

func (e Zvuk) Reauth(accountId, token string) error {
	return execTask(0, func(ctx context.Context) error {
		accById, err := accountByID(accountId)
		if err != nil {
			return err
		}
		return zvuk.Reauth(context.Background(), accById, token)
	})
}
