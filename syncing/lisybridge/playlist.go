package lisybridge

// import (
// 	"context"
// 	"errors"

// 	"github.com/oklookat/synchro/linking/linker"
// 	"github.com/oklookat/synchro/linking/linkerimpl"
// 	"github.com/oklookat/synchro/repository"
// 	"github.com/oklookat/synchro/shared"
// 	"github.com/oklookat/synchro/syncing/syncer"
// 	"github.com/oklookat/synchro/syncing/syncerimpl"
// )

// var (
// 	_syncablesPlaylists                                     = repository.SyncablesPlaylist{}
// 	_syncablesPlaylistName        syncer.Repository[string] = repository.SyncablesPlaylistName{}
// 	_syncablesPlaylistDescription syncer.Repository[string] = repository.SyncablesPlaylistDescription{}
// 	_syncablesPlaylistVisibility  syncer.Repository[bool]   = repository.SyncablesPlaylistIsVisible{}
// )

// func syncPlaylists(ctx context.Context, accounts map[shared.RepositoryID]*fullAccount) error {
// 	var pureAccounts []shared.Account
// 	for id := range accounts {
// 		pureAccounts = append(pureAccounts, accounts[id].Account)
// 	}
// 	lnk, err := linkerimpl.NewPlaylists(pureAccounts)
// 	if err != nil {
// 		return err
// 	}

// 	var playlistAccounts []*syncerimpl.PlaylistAccount

// 	// Create links.
// 	for _, acc := range accounts {
// 		playlists, err := acc.Actions.Playlist().MyPlaylists(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		_, err = linkPlaylists(ctx, lnk, acc.Account.ID(), playlists)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// Add accounts.
// 	for id, acc := range accounts {
// 		playlists, err := acc.Actions.Playlist().MyPlaylists(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		linkedPlaylists, err := linkPlaylists(ctx, lnk, acc.Account.ID(), playlists)
// 		if err != nil {
// 			return err
// 		}
// 		if err := onGotPlaylists(ctx, acc.Account, playlists); err != nil {
// 			return err
// 		}
// 		accounts[id].LinkedPlaylists = linkedPlaylists
// 		convertedLinkedPlaylists := map[shared.EntityID]shared.RemoteID{}
// 		for ei, rp := range linkedPlaylists {
// 			convertedLinkedPlaylists[ei] = rp.ID()
// 		}

// 		setts := acc.Settings.Playlists()
// 		wrapAcc := syncerimpl.NewPlaylistAccount(
// 			_syncablesPlaylists,
// 			lnk,
// 			setts.LastSynchronization(),
// 			setts.SetLastSynchronization,
// 			acc.Account.ID(),
// 			convertedLinkedPlaylists,
// 			acc.Actions.Playlist(),
// 		)
// 		playlistAccounts = append(playlistAccounts, wrapAcc)
// 	}

// 	// Self (delete).
// 	if err := sync2stages(func() error {
// 		for _, acc := range playlistAccounts {
// 			if err := acc.Start(ctx); err != nil {
// 				return err
// 			}
// 			for eId := range acc.NewRemoved {
// 				fAcc, ok := accounts[acc.AccountID]
// 				if !ok {
// 					continue
// 				}
// 				delete(fAcc.LinkedPlaylists, eId)
// 			}
// 		}
// 		return nil
// 	}); err != nil {
// 		return err
// 	}

// 	// Delete entities & links
// 	// (without this, deleted playlists links will remain in DB).
// 	removeLinks := map[shared.EntityID]bool{}
// 	for _, acc := range playlistAccounts {
// 		for eId := range acc.RemoveLinks {
// 			removeLinks[eId] = true
// 		}
// 	}
// 	if len(removeLinks) > 0 {
// 		if err := lnk.RemoveLinks(context.Background(), removeLinks); err != nil {
// 			return err
// 		}
// 	}
// 	if err := _syncablesPlaylists.DeleteUnsynced(); err != nil {
// 		return err
// 	}

// 	// Names.
// 	if err := syncPlaylistsNames(ctx, accounts); err != nil {
// 		return err
// 	}

// 	// Desc.
// 	if err := syncPlaylistsDescription(ctx, accounts); err != nil {
// 		return err
// 	}

// 	// Vis.
// 	if err := syncPlaylistsVisibility(ctx, accounts); err != nil {
// 		return err
// 	}

// 	// Tracks.
// 	return syncPlaylistsTracks(ctx, accounts)
// }

// func syncPlaylistsNames(
// 	ctx context.Context,
// 	accounts map[string]*fullAccount,
// ) error {

// 	var converted []MetaAccount[string]

// 	for accId := range accounts {
// 		nameSyncSettings := map[shared.EntityID]shared.SynchronizationSettings{}
// 		playlists := accounts[accId].LinkedPlaylists
// 		for id := range playlists {
// 			settings, err := accounts[accId].Settings.Playlist(id)
// 			if err != nil {
// 				return err
// 			}
// 			nameSyncSettings[id] = settings.Name()
// 		}
// 		converted = append(converted, playlistNameMetaAccount{
// 			nameSyncSettings: nameSyncSettings,
// 			playlists:        playlists,
// 		})
// 	}

// 	return sync2stages(func() error {
// 		return NewMetaSyncer(_syncablesPlaylistName, converted).Sync(ctx)
// 	})
// }

// func syncPlaylistsDescription(
// 	ctx context.Context,
// 	accounts map[string]*fullAccount,
// ) error {

// 	var converted []MetaAccount[string]

// 	for accId := range accounts {
// 		descSyncSettings := map[shared.EntityID]shared.SynchronizationSettings{}
// 		playlists := accounts[accId].LinkedPlaylists
// 		for id := range playlists {
// 			settings, err := accounts[accId].Settings.Playlist(id)
// 			if err != nil {
// 				return err
// 			}
// 			descSyncSettings[id] = settings.Description()
// 		}
// 		converted = append(converted, playlistDescriptionMetaAccount{
// 			descriptionSyncSettings: descSyncSettings,
// 			playlists:               playlists,
// 		})
// 	}

// 	return sync2stages(func() error {
// 		return NewMetaSyncer(_syncablesPlaylistDescription, converted).Sync(ctx)
// 	})
// }

// func syncPlaylistsVisibility(
// 	ctx context.Context,
// 	accounts map[string]*fullAccount,
// ) error {

// 	var converted []MetaAccount[bool]

// 	for accId := range accounts {
// 		visSyncSettings := map[shared.EntityID]shared.SynchronizationSettings{}
// 		playlists := accounts[accId].LinkedPlaylists
// 		for id := range playlists {
// 			settings, err := accounts[accId].Settings.Playlist(id)
// 			if err != nil {
// 				return err
// 			}
// 			visSyncSettings[id] = settings.Visibility()
// 		}
// 		converted = append(converted, playlistVisibilityMetaAccount{
// 			visibilitySyncSettings: visSyncSettings,
// 			playlists:              playlists,
// 		})
// 	}

// 	return sync2stages(func() error {
// 		return NewMetaSyncer(_syncablesPlaylistVisibility, converted).Sync(ctx)
// 	})
// }

// func syncPlaylistsTracks(ctx context.Context,
// 	accounts map[string]*fullAccount) error {

// 	lnk, err := linkerimpl.NewTracks()
// 	if err != nil {
// 		return err
// 	}

// 	var likeableAccounts []*syncerimpl.LikeableAccount

// 	for _, acc := range accounts {
// 		for eId := range acc.LinkedPlaylists {
// 			plSetts, err := acc.Settings.Playlist(eId)
// 			if err != nil {
// 				return err
// 			}
// 			if !plSetts.Tracks().Synchronize() {
// 				continue
// 			}

// 			repo, err := repository.NewSyncablesPlaylistTrack(ctx, eId)
// 			if err != nil {
// 				return err
// 			}

// 			tracks, err := acc.LinkedPlaylists[eId].Tracks(ctx)
// 			if err != nil {
// 				return err
// 			}
// 			tracksConverted := map[shared.RemoteID]shared.RemoteEntity{}
// 			for ri, rt := range tracks {
// 				tracksConverted[ri] = rt
// 			}
// 			tracksLinked, err := linkStatic(ctx, lnk, tracksConverted)
// 			if err != nil {
// 				return err
// 			}

// 			wrapAcc := syncerimpl.NewPlaylistTracksAccount(
// 				repo,
// 				lnk,
// 				acc.RemoteName,
// 				acc.LinkedPlaylists[eId],
// 				tracksLinked,
// 				plSetts.Tracks().LastSynchronization(),
// 				plSetts.Tracks().SetLastSynchronization,
// 			)
// 			likeableAccounts = append(likeableAccounts, wrapAcc)
// 		}
// 	}

// 	if err := sync2stages(func() error {
// 		for _, acc := range likeableAccounts {
// 			if err := acc.Start(ctx); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	}); err != nil {
// 		return err
// 	}

// 	for _, la := range likeableAccounts {
// 		la.Repo.DeleteUnsynced()
// 	}

// 	return err
// }

// type playlistNameMetaAccount struct {
// 	nameSyncSettings map[shared.EntityID]shared.SynchronizationSettings
// 	playlists        map[shared.EntityID]shared.RemotePlaylist
// }

// func (e playlistNameMetaAccount) MetaThings() []MetaThing[string] {
// 	var converted []MetaThing[string]
// 	for id := range e.playlists {
// 		_, ok := e.nameSyncSettings[id]
// 		if !ok {
// 			_log.Warn("_, ok := e.nameSyncSettings[id]")
// 			continue
// 		}
// 		converted = append(converted, syncerimpl.NewMetaString(
// 			id,
// 			e.playlists[id].Name(),
// 			e.playlists[id].Rename,
// 			e.nameSyncSettings[id],
// 		))
// 	}
// 	return converted
// }

// type playlistDescriptionMetaAccount struct {
// 	descriptionSyncSettings map[shared.EntityID]shared.SynchronizationSettings
// 	playlists               map[shared.EntityID]shared.RemotePlaylist
// }

// func (e playlistDescriptionMetaAccount) MetaThings() []MetaThing[string] {
// 	var converted []MetaThing[string]
// 	for id := range e.playlists {
// 		_, ok := e.descriptionSyncSettings[id]
// 		if !ok {
// 			_log.Warn("_, ok := e.descriptionSyncSettings[id]")
// 			continue
// 		}
// 		if e.playlists[id].Description() == nil {
// 			continue
// 		}
// 		converted = append(converted, syncerimpl.NewMetaString(
// 			id,
// 			*e.playlists[id].Description(),
// 			e.playlists[id].SetDescription,
// 			e.descriptionSyncSettings[id],
// 		))
// 	}
// 	return converted
// }

// type playlistVisibilityMetaAccount struct {
// 	visibilitySyncSettings map[shared.EntityID]shared.SynchronizationSettings
// 	playlists              map[shared.EntityID]shared.RemotePlaylist
// }

// func (e playlistVisibilityMetaAccount) MetaThings() []MetaThing[bool] {
// 	var converted []MetaThing[bool]
// 	for id := range e.playlists {
// 		_, ok := e.visibilitySyncSettings[id]
// 		if !ok {
// 			_log.Warn("_, ok := e.visibilitySyncSettings[id]")
// 			continue
// 		}
// 		vis, err := e.playlists[id].IsVisible()
// 		if err != nil {
// 			continue
// 		}
// 		converted = append(converted, syncerimpl.NewMetaBool(
// 			id,
// 			vis,
// 			e.playlists[id].SetIsVisible,
// 			e.visibilitySyncSettings[id],
// 		))
// 	}
// 	return converted
// }

// // Convert playlists to linked for account.
// //
// // Warning: playlists may be created for accounts by the linker.
// func linkPlaylists(
// 	ctx context.Context,
// 	lnk *linker.Dynamic,
// 	accountID string,
// 	playlists map[shared.RemoteID]shared.RemotePlaylist,
// ) (map[shared.EntityID]shared.RemotePlaylist, error) {
// 	result := make(map[shared.EntityID]shared.RemotePlaylist, len(playlists))
// 	for id := range playlists {
// 		linked, entity, err := lnk.FromRemote(
// 			ctx,
// 			linkerimpl.NewPlaylistRemoteEntity(playlists[id], accountID),
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if shared.IsNil(linked) {
// 			continue
// 		}
// 		conv, ok := entity.(*linkerimpl.PlaylistRemoteEntity)
// 		if !ok {
// 			return nil, errors.New("conv, ok := entity.(*linkerimpl.PlaylistRemoteEntity)")
// 		}
// 		result[linked.EntityID()] = conv.Real
// 	}
// 	return result, nil
// }
