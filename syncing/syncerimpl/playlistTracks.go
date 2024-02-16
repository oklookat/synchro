package syncerimpl

import (
	"context"
	"time"

	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/syncer"
)

type (
	TracksAdderRemover interface {
		AddTracks(context.Context, []shared.RemoteID) error
		RemoveTracks(context.Context, []shared.RemoteID) error
	}
)

func NewPlaylistTracksAccount(
	repo syncer.Repository[bool],
	lnk *linker.Static,

	remoteName shared.RemoteName,
	actions TracksAdderRemover,
	linkedPlaylistTracks map[shared.EntityID]shared.RemoteID,
	lastSync time.Time,
	setLastSync func(time.Time) error,
) *LikeableAccount {
	return NewLikeableAccount(
		repo,
		lnk,
		remoteName,
		linkedPlaylistTracks,
		lastSync,
		setLastSync,
		playlistTracksLikeUnliker{
			actions: actions,
		},
	)
}

type playlistTracksLikeUnliker struct {
	actions TracksAdderRemover
}

// Examples: like artists.
func (e playlistTracksLikeUnliker) Like(ctx context.Context, ids []shared.RemoteID) error {
	return e.actions.AddTracks(ctx, ids)
}

// Examples: unlike artists.
func (e playlistTracksLikeUnliker) Unlike(ctx context.Context, ids []shared.RemoteID) error {
	return e.actions.RemoveTracks(ctx, ids)
}
