package cli

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/oklookat/synchro/remote/deezer"
	"github.com/oklookat/synchro/remote/spotify"
	"github.com/oklookat/synchro/remote/vkmusic"
	"github.com/oklookat/synchro/remote/yandexmusic"
	"github.com/oklookat/synchro/remote/zvuk"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/vkmauth"
	"github.com/urfave/cli/v2"
)

type account struct {
}

func (e account) command() *cli.Command {
	return &cli.Command{
		Name:    "account",
		Aliases: []string{"a", "acc"},
		Subcommands: []*cli.Command{
			e.add(),
			e.list(),
			e.delete(),
			e.changeAlias(),
			e.reAuth(),
		},
		Usage: "Account(s) actions",
	}
}

func (a account) list() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l", "ls"},
		Usage:   "Show all accounts",
		Action: func(ctx *cli.Context) error {
			for _, rem := range repository.Remotes {
				accs, err := rem.Repository().Accounts(context.Background())
				if err != nil {
					return err
				}
				for _, acc := range accs {
					fmt.Printf("Remote: %s | ID: %s | Alias: %s", acc.RemoteName(), acc.ID(), acc.Alias())
				}
			}
			return nil
		},
	}
}

func (a account) delete() *cli.Command {
	return &cli.Command{
		Name:    "delete",
		Aliases: []string{"d", "del"},
		Usage:   "Delete account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Value:    "",
				Required: true,
				Usage:    "Account id",
			},
		},
		Action: func(ctx *cli.Context) error {
			id := ctx.String("id")
			acc, err := repository.AccountByID(shared.RepositoryID(id))
			if err != nil {
				return err
			}
			if shared.IsNil(acc) {
				slog.Error("Account not exists")
				return nil
			}
			return acc.Delete()
		},
	}
}

func (a account) changeAlias() *cli.Command {
	return &cli.Command{
		Name:    "alias",
		Aliases: []string{"al"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Value:    "",
				Required: true,
				Usage:    "Account id",
			},
			&cli.StringFlag{
				Name:     "alias",
				Value:    "",
				Required: true,
				Usage:    "New alias",
			},
		},
		Usage: "Change account alias",
		Action: func(ctx *cli.Context) error {
			id := ctx.String("id")
			alias := ctx.String("alias")
			acc, err := repository.AccountByID(shared.RepositoryID(id))
			if err != nil {
				return err
			}
			if shared.IsNil(acc) {
				slog.Error("Account not exists")
				return nil
			}
			return acc.SetAlias(alias)
		},
	}
}

func (a account) reAuth() *cli.Command {
	return &cli.Command{
		Name:    "reauth",
		Aliases: []string{"re"},
		Usage:   "Reauth on account",
	}
}

func (e account) add() *cli.Command {
	_idSecretFlags := []cli.Flag{
		&cli.StringFlag{
			Required: true,
			Name:     "id",
			Usage:    "App ID",
			Value:    "",
		},
		&cli.StringFlag{
			Required: true,
			Name:     "secret",
			Usage:    "App Secret",
			Value:    "",
		},
	}

	return &cli.Command{
		Name:    "add",
		Aliases: []string{"a"},
		Usage:   "Add streaming service account",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "alias",
				Value:   "",
				Usage:   "Account alias (optional)",
				Aliases: []string{"a"},
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:    "deezer",
				Aliases: []string{"de"},
				Usage:   "Add Deezer account",
				Flags:   _idSecretFlags,
				Action: func(ctx *cli.Context) error {
					alias := ctx.String("alias")
					id := ctx.String("id")
					secret := ctx.String("secret")
					acc, err := deezer.NewAccount(context.Background(), alias, id, secret, func(url string) {
						slog.Info("Go to", "URL", url)
						shared.OpenBrowser(url)
					})
					if err != nil {
						return err
					}
					e.onAccountCreated(acc)
					return err
				},
			},
			{
				Name:    "spotify",
				Aliases: []string{"sp"},
				Usage:   "Add Spotify account",
				Flags:   _idSecretFlags,
				Action: func(ctx *cli.Context) error {
					alias := ctx.String("alias")
					id := ctx.String("id")
					secret := ctx.String("secret")
					acc, err := spotify.NewAccount(context.Background(), alias, id, secret, func(url string) {
						slog.Info("Go to", "URL", url)
						shared.OpenBrowser(url)
					})
					if err != nil {
						return err
					}
					e.onAccountCreated(acc)
					return err
				},
			},
			{
				Name:    "vkmusic",
				Aliases: []string{"vkm"},
				Usage:   "Add VK Music account",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Required: true,
						Name:     "phone",
						Value:    "",
						Usage:    "VK phone number",
					},
					&cli.StringFlag{
						Required: true,
						Name:     "password",
						Value:    "",
						Usage:    "VK password",
					},
				},
				Action: func(ctx *cli.Context) error {
					alias := ctx.String("alias")
					phone := ctx.String("phone")
					password := ctx.String("password")
					acc, err := vkmusic.NewAccount(context.Background(), alias, phone, password,
						func(by vkmauth.CodeSended) (vkmauth.GotCode, error) {
							for {
								slog.Info("Code sended", "TO", by.Current.String())
								canResend := len(by.Resend) > 0
								enterMsg := "Enter code"
								if canResend {
									slog.Info("Code also can be sended", "TO", by.Resend.String())
									enterMsg += " or leave blank for resend code"
								}
								slog.Info(enterMsg)
								code, err := readInput()
								if err != nil {
									return vkmauth.GotCode{}, err
								}
								if len(code) < 1 {
									if canResend {
										return vkmauth.GotCode{Resend: true}, nil
									}
									slog.Error("Code cant be resended")
									continue
								}
								return vkmauth.GotCode{Code: code}, nil
							}
						})
					if err != nil {
						return err
					}
					e.onAccountCreated(acc)
					return err
				},
			},
			{
				Name:    "yandexmusic",
				Aliases: []string{"ym"},
				Usage:   "Add Yandex.Music account",
				Action: func(ctx *cli.Context) error {
					alias := ctx.String("alias")
					acc, err := yandexmusic.NewAccount(context.Background(), alias, func(url, code string) {
						slog.Info("Go to", "URL", url)
						slog.Info("And enter", "CODE", code)
						shared.OpenBrowser(url)
					})
					if err != nil {
						return err
					}
					e.onAccountCreated(acc)
					return err
				},
			},
			{
				Name:  "zvuk",
				Usage: "Add Zvuk account",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Required: true,
						Name:     "token",
						Usage:    "Token from https://zvuk.com/api/tiny/profile",
					},
				},
				Action: func(ctx *cli.Context) error {
					alias := ctx.String("alias")
					token := ctx.String("token")
					acc, err := zvuk.NewAccount(context.Background(), alias, token)
					if err != nil {
						return err
					}
					e.onAccountCreated(acc)
					return err
				},
			},
		},
	}
}

func (e account) onAccountCreated(acc shared.Account) {
	slog.Info("Account created", "ID", acc.ID().String(), "remote", acc.RemoteName().String(), "alias", acc.Alias())
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	return reader.ReadString('\n')
}
