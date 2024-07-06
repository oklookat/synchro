package vkmusic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/oklookat/govkm"
	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/vkmauth"

	"golang.org/x/oauth2"
)

var (
	errNilPlaylist = errors.New("nil playlist")
	errNilAlbum    = errors.New("nil album")
)

func NewAccount(
	ctx context.Context,
	alias string,
	phone string,
	password string,
	onCodeWaiting func(by vkmauth.CodeSended) (vkmauth.GotCode, error),
) (shared.Account, error) {

	token, err := vkmauth.New(ctx, phone, password, onCodeWaiting)
	if err != nil {
		return nil, err
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(alias, string(tokenBytes))
	if err != nil {
		return nil, err
	}

	return account, err
}

func Reauth(
	ctx context.Context,
	account shared.Account,
	phone string,
	password string,
	onCodeWaiting func(by vkmauth.CodeSended) (vkmauth.GotCode, error),
) error {
	token, err := vkmauth.New(ctx, phone, password, onCodeWaiting)
	if err != nil {
		return err
	}
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return account.SetAuth(string(tokenBytes))
}

func getClient(account shared.Account) (*govkm.Client, error) {
	token := &oauth2.Token{}
	if err := json.Unmarshal([]byte(account.Auth()), token); err != nil {
		return nil, err
	}
	hClient, err := getHttpProxyClient()
	if err != nil {
		return nil, err
	}
	// TODO: тут и еще много где надо сразу клиент передавать. И еще надо отказаться от vantuz
	// еще можно конфиги для стримингов создавать автоматически, и получать конфиги прокси для них тоже
	// т.е инстанс конфига можно привязать к Remote, типа как репозиторий к нему привязывается
	cl, err := govkm.New(token.AccessToken)
	if err != nil {
		return nil, err
	}
	cl.Http.SetClient(hClient)
	cl.Http.SetRateLimit(2, time.Second)
	return cl, err
}

func getHttpProxyClient() (*http.Client, error) {
	cfg, err := config.Get[*config.VKMusic](config.KeyVKMusic)
	if err != nil {
		return nil, err
	}

	// Proxy?
	hClient := &http.Client{}
	if (*cfg).Proxy.Proxy {
		pUrl, err := url.Parse((*cfg).Proxy.URL)
		if err != nil {
			return nil, err
		}
		hClient.Transport = &http.Transport{Proxy: http.ProxyURL(pUrl)}
	}

	return hClient, err
}
