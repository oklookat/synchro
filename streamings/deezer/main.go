package deezer

import (
	"context"
	"net/http"
	"time"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/deezerauth"
	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/streaming"
	"golang.org/x/oauth2"
)

var (
	_state = "abc123"
)

func NewAccount(
	ctx context.Context,
	alias string,
	appID string,
	appSecret string,
	onURL func(url string),
) (streaming.Account, error) {

	token, err := getToken(ctx, appID, appSecret, onURL)
	if err != nil {
		return nil, err
	}

	tokenStr, err := shared.TokenToAuth(token)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(alias, tokenStr)
	if err != nil {
		return nil, err
	}

	return account, err
}

func getToken(ctx context.Context, appID, appSecret string, onURL func(url string)) (*oauth2.Token, error) {
	args, err := getArgs(appID, appSecret, onURL)
	if err != nil {
		return nil, err
	}
	return deezerauth.New(ctx, args)
}

func getArgs(appID, appSecret string, onURL func(url string)) (deezerauth.AuthArgs, error) {
	cfg := &Config{}
	if err := config.Get(cfg.Key(), cfg); err != nil {
		return deezerauth.AuthArgs{}, err
	}

	perms := []deezerauth.Permission{
		deezerauth.PermissionBasicAccess,
		deezerauth.PermissionEmail,
		deezerauth.PermissionOfflineAccess,
		deezerauth.PermissionManageLibrary,
		deezerauth.PermissionManageCommunity,
		deezerauth.PermissionDeleteLibrary,
		deezerauth.PermissionListeningHistory,
	}
	return deezerauth.AuthArgs{
		State:       _state,
		AppID:       appID,
		Secret:      appSecret,
		RedirectUri: cfg.Host,
		Port:        cfg.Port,
		Perms:       perms,
		OnURL:       onURL,
	}, nil
}

func getClient(account streaming.Account) (*deezus.Client, error) {
	token, err := shared.AuthToToken(account.Auth())
	if err != nil {
		return nil, err
	}
	cl, err := deezus.New(token.AccessToken)
	if err != nil {
		return nil, err
	}
	cl.Http.SetClient(&http.Client{
		Timeout: 15 * time.Second,
	})
	return cl, err
}
