package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/urfave/cli/v2"
)

type destruct struct {
}

func (e destruct) command() *cli.Command {
	return &cli.Command{
		Name:    "destruct",
		Aliases: []string{"d"},
		Subcommands: []*cli.Command{
			e.accountThings(),
		},
		Usage: "Destructive things",
	}
}

func (e destruct) accountThings() *cli.Command {
	return &cli.Command{
		Name:    "account",
		Aliases: []string{"a"},
		Usage:   "Destructive account things",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Value:    "",
				Required: true,
				Usage:    "Account id",
			},
			&cli.BoolFlag{
				Name:     "likedAlbums",
				Aliases:  []string{"lab"},
				Value:    false,
				Required: false,
				Usage:    "Delete all liked albums?",
			},
			&cli.BoolFlag{
				Name:     "likedArtists",
				Aliases:  []string{"lar"},
				Value:    false,
				Required: false,
				Usage:    "Delete all liked artists?",
			},
			&cli.BoolFlag{
				Name:     "likedTracks",
				Aliases:  []string{"ltr"},
				Value:    false,
				Required: false,
				Usage:    "Delete all liked tracks?",
			},
			&cli.BoolFlag{
				Name:     "playlists",
				Aliases:  []string{"pla"},
				Value:    false,
				Required: false,
				Usage:    "Delete all playlists?",
			},
		},
		Action: func(ctx *cli.Context) error {
			from := ctx.String("id")
			acc, err := e.getAcc(from)
			if err != nil {
				return err
			}
			acts, err := acc.Actions()
			if err != nil {
				return err
			}
			if ctx.Bool("likedAlbums") {
				slog.Info("Deleting", "what", "liked albums")
				act := acts.LikedAlbums()
				if err := e.deleteEntities(act); err != nil {
					return err
				}
			}
			if ctx.Bool("likedArtists") {
				slog.Info("Deleting", "what", "liked artists")
				act := acts.LikedArtists()
				if err := e.deleteEntities(act); err != nil {
					return err
				}
			}
			if ctx.Bool("likedTracks") {
				slog.Info("Deleting", "what", "liked tracks")
				act := acts.LikedTracks()
				if err := e.deleteEntities(act); err != nil {
					return err
				}
			}
			if ctx.Bool("playlists") {
				slog.Info("Deleting", "what", "playlists")
				pls, err := acts.Playlist().MyPlaylists(context.Background())
				if err != nil {
					return err
				}
				ids := make([]shared.RemoteID, len(pls))
				for i := range pls {
					ids[i] = pls[i].ID()
				}
				if err := acts.Playlist().Delete(context.Background(), ids); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func (e destruct) getAcc(id string) (shared.Account, error) {
	acc, err := repository.AccountByID(shared.RepositoryID(id))
	if err != nil {
		return nil, err
	}
	if shared.IsNil(acc) {
		slog.Error("Account not exists")
		return nil, errors.New("account not exists")
	}
	return acc, err
}

func (e destruct) deleteEntities(acts shared.LikedActions) error {
	ents, err := acts.Liked(context.Background())
	if err != nil {
		return err
	}
	ids := make([]shared.RemoteID, len(ents))
	for i := range ents {
		ids[i] = ents[i].ID()
	}
	return acts.Unlike(context.Background(), ids)
}
