package cli

import (
	"context"
	"errors"
	"log/slog"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

type transfer struct {
}

func (e transfer) command() *cli.Command {
	return &cli.Command{
		Name:    "transfer",
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "from",
				Aliases:  []string{"f"},
				Value:    "",
				Required: true,
				Usage:    "From account id",
			},
			&cli.StringFlag{
				Name:     "to",
				Aliases:  []string{"t"},
				Value:    "",
				Required: true,
				Usage:    "To account id",
			},
			&cli.BoolFlag{
				Name:     "likedAlbums",
				Aliases:  []string{"lab"},
				Value:    true,
				Required: false,
				Usage:    "Liked albums?",
			},
			&cli.BoolFlag{
				Name:     "likedArtists",
				Aliases:  []string{"lar"},
				Value:    true,
				Required: false,
				Usage:    "Liked artists?",
			},
			&cli.BoolFlag{
				Name:     "likedTracks",
				Aliases:  []string{"ltr"},
				Value:    true,
				Required: false,
				Usage:    "Liked tracks?",
			},
			&cli.BoolFlag{
				Name:     "playlists",
				Aliases:  []string{"pla"},
				Value:    true,
				Required: false,
				Usage:    "Playlists?",
			},
		},
		Usage: "Transfer entities between accounts",
		Action: func(ctx *cli.Context) error {
			from := ctx.String("from")
			fromAcc, err := e.getAcc(from)
			if err != nil {
				return err
			}
			to := ctx.String("to")
			toAcc, err := e.getAcc(to)
			if err != nil {
				return err
			}

			if ctx.Bool("likedAlbums") {
				slog.Info("Transfering", "what", "liked albums")
				lnk, err := linkerimpl.NewAlbums()
				if err != nil {
					return err
				}
				fromActs, err := fromAcc.Actions()
				if err != nil {
					return err
				}
				fromAct := fromActs.LikedAlbums()
				toActs, err := toAcc.Actions()
				if err != nil {
					return err
				}
				toAct := toActs.LikedAlbums()
				if err := e.transferBtw(lnk, fromAcc, toAcc, fromAct, toAct); err != nil {
					return err
				}
			}

			if ctx.Bool("likedArtists") {
				slog.Info("Transfering", "what", "liked artists")
				lnk, err := linkerimpl.NewArtists()
				if err != nil {
					return err
				}
				fromActs, err := fromAcc.Actions()
				if err != nil {
					return err
				}
				fromAct := fromActs.LikedArtists()
				toActs, err := toAcc.Actions()
				if err != nil {
					return err
				}
				toAct := toActs.LikedArtists()
				if err := e.transferBtw(lnk, fromAcc, toAcc, fromAct, toAct); err != nil {
					return err
				}
			}

			if ctx.Bool("likedTracks") {
				slog.Info("Transfering", "what", "liked tracks")
				lnk, err := linkerimpl.NewTracks()
				if err != nil {
					return err
				}
				fromActs, err := fromAcc.Actions()
				if err != nil {
					return err
				}
				fromAct := fromActs.LikedTracks()
				toActs, err := toAcc.Actions()
				if err != nil {
					return err
				}
				toAct := toActs.LikedTracks()
				if err := e.transferBtw(lnk, fromAcc, toAcc, fromAct, toAct); err != nil {
					return err
				}
			}

			if ctx.Bool("playlists") {
				slog.Info("Transfering", "what", "playlists")
				lnk, err := linkerimpl.NewTracks()
				if err != nil {
					return err
				}
				fromActs, err := fromAcc.Actions()
				if err != nil {
					return err
				}
				fromAct := fromActs.Playlist()
				toActs, err := toAcc.Actions()
				if err != nil {
					return err
				}
				toAct := toActs.Playlist()

				rCtx := context.Background()

				fromPlaylists, err := fromAct.MyPlaylists(rCtx)
				if err != nil {
					return err
				}

				for _, fromPlaylist := range fromPlaylists {
					slog.Info("Current playlist", "Name", fromPlaylist.Name())

					isVis, _ := fromPlaylist.IsVisible()

					toPlaylist, err := toAct.Create(rCtx, fromPlaylist.Name(), isVis, fromPlaylist.Description())
					if err != nil {
						return err
					}

					fromWrapAct := playlistLikedActions{pl: fromPlaylist}
					toWrapAct := playlistLikedActions{pl: toPlaylist}

					if err := e.transferBtw(lnk, fromAcc, toAcc, fromWrapAct, toWrapAct); err != nil {
						toAct.Delete(rCtx, []shared.RemoteID{toPlaylist.ID()})
						return err
					}
				}
			}

			return nil
		},
	}
}

func (e transfer) transferBtw(
	lnk *linker.Static,
	fromAcc shared.Account, toAcc shared.Account,
	fromAct shared.LikedActions, toAct shared.LikedActions,
) error {
	ctx := context.Background()

	slog.Info("Transfer BTW",
		"from remote",
		fromAcc.RemoteName().String(),
		"to remote", toAcc.RemoteName().String(),
		"from account id", fromAcc.ID().String(),
		"to account id", toAcc.ID().String())

	slog.Info("Getting liked...", "account id", fromAcc.ID())
	liked, err := fromAct.Liked(ctx)
	if err != nil {
		return err
	}

	bar := progressbar.Default(int64(len(liked)))
	bar.Describe("Linking (Remote -> DB)")

	fromLinkedList := []linker.Linked{}
	for _, ent := range liked {
		linkedRes, err := lnk.FromRemote(ctx, ent)
		if err != nil {
			return err
		}
		fromLinkedList = append(fromLinkedList, linkedRes.Linked)
		bar.Add(1)
	}

	slog.Info("Why don't we have a cup of tea? 🤔")

	bar.Reset()
	bar.ChangeMax(len(fromLinkedList))
	bar.Describe("Linking (DB -> Remote)")

	toLinkedIds := []shared.RemoteID{}
	for i, linked := range fromLinkedList {
		if shared.IsNil(linked) {
			slog.Warn("Not found", "Name", liked[i].Name(), "ID", liked[i].ID().String())
			continue
		}
		res, err := lnk.ToRemote(ctx, linked.EntityID(), toAcc.RemoteName())
		if err != nil {
			return err
		}
		if res.MissingNow || shared.IsNil(res.Linked) || res.Linked.RemoteID() == nil {
			slog.Warn("Not found", "Name", liked[i].Name(), "ID", liked[i].ID().String())
			continue
		}
		toLinkedIds = append(toLinkedIds, *res.Linked.RemoteID())
	}
	bar.Exit()
	slog.Info("Liking", "entitiesCount", len(toLinkedIds))
	return toAct.Like(context.Background(), toLinkedIds)
}

func (e transfer) getAcc(id string) (shared.Account, error) {
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

type playlistLikedActions struct {
	pl shared.RemotePlaylist
}

func (e playlistLikedActions) Liked(ctx context.Context) ([]shared.RemoteEntity, error) {
	trs, err := e.pl.Tracks(ctx)
	if err != nil {
		return nil, err
	}
	ents := make([]shared.RemoteEntity, len(trs))
	for i := range trs {
		ents[i] = trs[i]
	}
	return ents, err
}

func (e playlistLikedActions) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.pl.AddTracks(ctx, ids)
}

func (e playlistLikedActions) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.pl.RemoveTracks(ctx, ids)
}
