package lisybridge

import (
	"context"

	"github.com/oklookat/synchro/shared"
)

var (
	_snapHook = &snapshotHooker{}
)

func onSyncStart(ctx context.Context) error {
	_snapHook = &snapshotHooker{}
	return nil
}

func onSyncEnd(ctx context.Context) error {
	if _snapHook != nil {
		_snapHook = nil
	}
	return nil
}

func onGotLikedAlbums(ctx context.Context, account shared.Account, entities map[shared.RemoteID]shared.RemoteEntity) error {
	return _snapHook.onLikedAlbums(ctx, account, entities)
}

func onGotLikedArtists(ctx context.Context, account shared.Account, entities map[shared.RemoteID]shared.RemoteEntity) error {
	return _snapHook.onLikedArtists(ctx, account, entities)
}

func onGotLikedTracks(ctx context.Context, account shared.Account, entities map[shared.RemoteID]shared.RemoteEntity) error {
	return _snapHook.onLikedTracks(ctx, account, entities)
}

func onGotPlaylists(ctx context.Context, account shared.Account, entities map[shared.RemoteID]shared.RemotePlaylist) error {
	return _snapHook.onPlaylists(ctx, account, entities)
}
