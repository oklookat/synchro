package deezer

import (
	"context"

	"github.com/oklookat/deezus"
	"github.com/oklookat/deezus/deezerauth"
	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
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
) (shared.Account, error) {

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

func Reauth(
	ctx context.Context,
	account shared.Account,
	appID string,
	appSecret string,
	onURL func(url string),
) error {
	token, err := getToken(ctx, appID, appSecret, onURL)
	if err != nil {
		return err
	}
	tokenStr, err := shared.TokenToAuth(token)
	if err != nil {
		return err
	}
	return account.SetAuth(tokenStr)
}

func getToken(ctx context.Context, appID, appSecret string, onURL func(url string)) (*oauth2.Token, error) {
	args, err := getArgs(appID, appSecret, onURL)
	if err != nil {
		return nil, err
	}
	return deezerauth.New(ctx, args)
}

func getArgs(appID, appSecret string, onURL func(url string)) (deezerauth.AuthArgs, error) {
	cfg, err := config.Get[config.Deezer](config.KeyDeezer)
	if err != nil {
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
	}, err
}

func getClient(account shared.Account) (*deezus.Client, error) {
	token, err := shared.AuthToToken(account.Auth())
	if err != nil {
		return nil, err
	}
	return deezus.New(token.AccessToken)
}
