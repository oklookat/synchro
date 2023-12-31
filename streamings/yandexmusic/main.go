package yandexmusic

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/oklookat/goym"
	"github.com/oklookat/yandexauth/v2"

	"golang.org/x/oauth2"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
)

var (
	errEmptyID = errors.New("empty ID")
)

const (
	// YM Windows app.
	_clientID     = "23cabbbdc6cd418abb4b39c32c41195d"
	_clientSecret = "53bc75238f0c4d08a118e51fe9203300"
)

func NewAccount(
	ctx context.Context,
	alias string,
	onUrlCode func(url string, code string),
) (streaming.Account, error) {
	tokens, err := getTokens(ctx, onUrlCode)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(alias, tokens)
	if err != nil {
		return nil, err
	}

	return account, err
}

func getTokens(ctx context.Context, onUrlCode func(url, code string)) (string, error) {
	cfg := &Config{}
	if err := config.Get(cfg.Key(), cfg); err != nil {
		return "", err
	}

	tokens, err := yandexauth.New(
		ctx,
		_clientID,
		_clientSecret,
		cfg.DeviceID,
		fmt.Sprintf("synchro (%s)", runtime.GOOS),
		onUrlCode,
	)

	if err != nil {
		return "", err
	}

	return shared.TokenToAuth(tokens)
}

func getClient(account streaming.Account) (*goym.Client, error) {
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
