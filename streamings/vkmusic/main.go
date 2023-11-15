package vkmusic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/oklookat/govkm"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
	"github.com/oklookat/vkmauth"

	"golang.org/x/oauth2"
)

var (
	errNilPlaylist = errors.New("nil playlist")
	errNilAlbum    = errors.New("nil album")
)

func NewAccount(
	ctx context.Context,
	alias *string,
	phone string,
	password string,
	onCodeWaiting func(by vkmauth.CodeSended) (vkmauth.GotCode, error),
) (streaming.Account, error) {

	token, err := vkmauth.New(ctx, phone, password, onCodeWaiting)
	if err != nil {
		return nil, err
	}

	if alias == nil || len(*alias) == 0 {
		randWord := shared.GenerateWord()
		alias = &randWord
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(*alias, string(tokenBytes))
	if err != nil {
		return nil, err
	}

	return account, err
}

func getClient(account streaming.Account) (*govkm.Client, error) {
	token := &oauth2.Token{}
	if err := json.Unmarshal([]byte(account.Auth()), token); err != nil {
		return nil, err
	}
	cl, err := govkm.New(token.AccessToken)
	if err != nil {
		return nil, err
	}
	cl.Http.SetRateLimit(2, time.Second)
	return cl, err
}
