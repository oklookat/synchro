package yandexmusic

import (
	"context"
	"errors"

	"github.com/oklookat/goym"
	"github.com/oklookat/yandexauth"
	"golang.org/x/oauth2"

	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/shared"
)

var (
	_log *logger.Logger

	errEmptyID = errors.New("empty ID")
)

const (
	// YM Windows app.

	_clientID     = "23cabbbdc6cd418abb4b39c32c41195d"
	_clientSecret = "53bc75238f0c4d08a118e51fe9203300"
)

func NewAccount(
	ctx context.Context,
	alias *string,
	login string,
	onUrlCode func(url string, code string),
) (shared.Account, error) {

	tokens, err := getTokens(ctx, login, onUrlCode)
	if err != nil {
		return nil, err
	}

	if alias == nil || len(*alias) == 0 {
		alias = &login
	}

	account, err := _repo.CreateAccount(*alias, tokens)
	if err != nil {
		return nil, err
	}

	return account, err
}

func Reauth(
	ctx context.Context,
	account shared.Account,
	login string,
	onUrlCode func(url string, code string),
) error {

	tokens, err := getTokens(ctx, login, onUrlCode)

	if err != nil {
		return err
	}

	return account.SetAuth(tokens)
}

func getTokens(ctx context.Context,
	login string,
	onUrlCode func(url, code string)) (string, error) {

	hostname := "synchro " + shared.GenerateWord()
	tokens, err := yandexauth.New(
		ctx,
		_clientID,
		_clientSecret,
		login,
		hostname,
		onUrlCode,
	)

	if err != nil {
		return "", err
	}

	return shared.TokenToAuth(tokens)
}

func getClient(account shared.Account) (*goym.Client, error) {
	tokens, err := shared.AuthToToken(account.Auth())
	if err != nil {
		return nil, err
	}

	// Refresh if needed.
	var refreshed *oauth2.Token
	if !tokens.Valid() {
		refreshed, err = yandexauth.Refresh(context.Background(), tokens.RefreshToken, _clientID, _clientSecret)
		if err != nil {
			return nil, err
		}
	}

	if refreshed != nil {
		au, err := shared.TokenToAuth(refreshed)
		if err != nil {
			return nil, err
		}
		if err := account.SetAuth(au); err != nil {
			return nil, err
		}
		tokens = refreshed
	}

	// Create client.
	return goym.New(tokens.AccessToken)
}
