package yandexmusic

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/oklookat/goym"
	"github.com/oklookat/yandexauth/v2"
	"golang.org/x/oauth2"

	"github.com/oklookat/synchro/shared"
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
) (shared.Account, error) {

	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 || hostname == "localhost" {
		hostname = "synchro " + shared.GenerateWord()
	}
	deviceID := shared.GenerateULID()

	tokens, err := getTokens(ctx, deviceID, hostname, onUrlCode)
	if err != nil {
		return nil, err
	}

	var tokFinal theToken
	tokFinal.Token = tokens
	tokFinal.DeviceID = deviceID
	tokFinal.Hostname = hostname

	jBytes, err := json.Marshal(&tokFinal)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(alias, string(jBytes))
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
	var tok theToken
	if err := json.Unmarshal([]byte(account.Auth()), &tok); err != nil {
		return err
	}

	tokens, err := getTokens(ctx, tok.DeviceID, tok.Hostname, onUrlCode)
	if err != nil {
		return err
	}

	tok.Token = tokens

	jBytes, err := json.Marshal(&tok)
	if err != nil {
		return err
	}

	return account.SetAuth(string(jBytes))
}

func getTokens(
	ctx context.Context,
	deviceID, hostname string,
	onUrlCode func(url, code string),
) (*oauth2.Token, error) {
	return yandexauth.New(ctx, _clientID, _clientSecret, deviceID, hostname, onUrlCode)
}

type theToken struct {
	*oauth2.Token
	DeviceID string `json:"deviceID"`
	Hostname string `json:"hostname"`
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
