package zvuk

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oklookat/gozvuk"

	"github.com/oklookat/synchro/shared"
)

func NewAccount(ctx context.Context, alias string, token string) (shared.Account, error) {
	token = strings.TrimSpace(token)
	tok := newToken(token)
	auth, err := shared.TokenToAuth(tok)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(alias, auth)
	if err != nil {
		return nil, err
	}

	return account, err
}

func Reauth(ctx context.Context,
	account shared.Account,
	accessToken string) error {

	tok := newToken(accessToken)
	auth, err := shared.TokenToAuth(tok)
	if err != nil {
		return err
	}

	return account.SetAuth(auth)
}

func getClient(account shared.Account) (*gozvuk.Client, error) {
	tok, err := shared.AuthToToken(account.Auth())
	if err != nil {
		return nil, err
	}

	// Create client.
	client := gozvuk.New(tok.AccessToken)

	_, err = client.Profile()
	if err != nil {
		return nil, errors.New("ping: " + err.Error())
	}

	client.Http.SetRateLimit(10, time.Second)

	return client, err
}
