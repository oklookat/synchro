package spotify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	_state = "abc123"

	// So far, there have been no problems with AU, for example when searching.
	_market = spotify.Market("AU")
)

func NewAccount(ctx context.Context,
	alias string,
	clientID string,
	clientSecret string,
	onURL func(url string)) (shared.Account, error) {

	tokens, err := getTokens(ctx, clientID, clientSecret, onURL)
	if err != nil {
		return nil, err
	}

	account, err := _repo.CreateAccount(alias, tokens)
	if err != nil {
		return nil, err
	}

	return account, err
}

func Reauth(
	ctx context.Context,
	account shared.Account,
	clientID string,
	clientSecret string,
	onURL func(url string)) error {
	tokens, err := getTokens(ctx, clientID, clientSecret, onURL)
	if err != nil {
		return err
	}
	return account.SetAuth(tokens)
}

func getTokens(ctx context.Context, clientID string, clientSecret string, onURL func(url string)) (string, error) {
	httpErr := make(chan error)
	auth, err := getAuthenticator(clientID, clientSecret)
	if err != nil {
		return "", err
	}
	clientCh := make(chan *spotify.Client)

	go serve(ctx, func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), _state, r)
		if err != nil {
			httpErr <- err
			return
		}

		if st := r.FormValue("state"); st != _state {
			msg := fmt.Errorf("state mismatch. Actual: %s, got: %s", st, _state)
			w.WriteHeader(404)
			w.Write([]byte(msg.Error()))
			httpErr <- msg
			return
		}

		auClient := auth.Client(r.Context(), tok)
		if err = setProxyClient(auClient); err != nil {
			httpErr <- err
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		clientCh <- spotify.New(auClient, spotify.WithRetry(true))
		w.WriteHeader(200)
		w.Write([]byte("Done. Now you can go back to where you came from."))
		httpErr <- err
	})

	// get auth url
	url := auth.AuthURL(_state)

	// send url to user
	go onURL(url)

	var client *spotify.Client
L:
	for {
		select {
		// check err from http handler
		case err := <-httpErr:
			if err != nil {
				return "", err
			}
		// check client from handler
		case client = <-clientCh:
			if client == nil {
				return "", errors.New("nil client")
			}
			break L
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// get tokens
	token, err := client.Token()
	if err != nil {
		return "", err
	}

	// create remote acc
	au := &authorized{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Token:        token,
	}

	rized, err := authorizedToAuth(au)
	if err != nil {
		return "", err
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		return "", context.Canceled
	}

	return rized, err
}

type authorized struct {
	ClientID     string        `json:"clientID"`
	ClientSecret string        `json:"clientSecret"`
	Token        *oauth2.Token `json:"token"`
}

func getAuthenticator(clientID, clientSecret string) (*spotifyauth.Authenticator, error) {
	cfg, err := config.Get[*config.Spotify](config.KeySpotify)
	if err != nil {
		return nil, err
	}

	fullUrl := (*cfg).Host + ":" + strconv.Itoa(int((*cfg).Port))
	return spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(fullUrl),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,

			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserLibraryModify,

			spotifyauth.ScopeUserFollowRead,
			spotifyauth.ScopeUserFollowModify,

			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
		),
	), err
}

func serve(ctx context.Context, what http.HandlerFunc) (err error) {
	cfg, err := config.Get[*config.Spotify](config.KeySpotify)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", what)

	port := ":" + strconv.Itoa(int((*cfg).Port))
	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen: " + err.Error())
		}
	}()

	slog.Debug("server started")

	<-ctx.Done()

	slog.Debug("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		if !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
			slog.Error("shutdown: " + err.Error())
		}
	}

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

func getClient(account shared.Account) (*spotify.Client, error) {
	token, err := authToAuthorized(account.Auth())
	if err != nil {
		return nil, err
	}
	au, err := getAuthenticator(token.ClientID, token.ClientSecret)
	if err != nil {
		return nil, err
	}
	auClient := au.Client(context.Background(), token.Token)
	if err := setProxyClient(auClient); err != nil {
		return nil, err
	}
	client := spotify.New(auClient)
	return client, err
}

func setProxyClient(fromClient *http.Client) error {
	cfg, err := config.Get[*config.Spotify](config.KeySpotify)
	if err != nil {
		return err
	}

	// Proxy?
	if (*cfg).Proxy.Proxy {
		pUrl, err := url.Parse((*cfg).Proxy.URL)
		if err != nil {
			return err
		}
		trs, ok := fromClient.Transport.(*oauth2.Transport)
		if !ok {
			return errors.New("fromClient.Transport.(*oauth2.Transport)")
		}
		trs.Base = &http.Transport{Proxy: http.ProxyURL(pUrl)}
	}

	return err
}
