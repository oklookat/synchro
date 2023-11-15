package tools

import (
	"context"

	"github.com/oklookat/synchro/streaming"
)

func UnlikeLiked(ctx context.Context, acts streaming.LikedActions) error {
	liked, err := acts.Liked(ctx)
	if err != nil {
		return err
	}
	ids := make([]streaming.ServiceEntityID, len(liked))
	for i := range ids {
		ids[i] = liked[i].ID()
	}
	return acts.Unlike(ctx, ids)
}

func DeleteAllPlaylists(ctx context.Context, acts streaming.PlaylistActions) error {
	playlists, err := acts.MyPlaylists(ctx)
	if err != nil {
		return err
	}
	ids := make([]streaming.ServiceEntityID, len(playlists))
	for i := range ids {
		ids[i] = playlists[i].ID()
	}
	return acts.Delete(ctx, ids)
}
