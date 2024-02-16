package commander

import (
	"context"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/shared"
)

func NewConfigLinker() *ConfigLinker {
	cfg := &config.Linker{}
	if err := config.Get(cfg); err != nil {
		cfg.Default()
	}
	return &ConfigLinker{
		self: cfg,
	}
}

type ConfigLinker struct {
	self *config.Linker
}

func (e *ConfigLinker) RecheckMissing() bool {
	return e.self.RecheckMissing
}

func (e *ConfigLinker) SetRecheckMissing(val bool) error {
	e.self.RecheckMissing = val
	return config.Save(e.self)
}

// Get URL of entity.
func LinkerEntityURL(remoteName, remoteID, entityType string) (string, error) {
	var remName shared.RemoteName
	remName.FromString(remoteName)
	var remID shared.RemoteID
	remID.FromString(remoteID)
	var entType shared.EntityType
	entType.FromString(entityType)
	rem, ok := _remotes[remName]
	if !ok {
		return "", shared.NewErrRemoteNotFound(_packageName, remName)
	}
	urld := rem.EntityURL(entType, remID)
	return urld.String(), nil
}

// Albums links.
func LinkerLinksAlbums(remoteName string) (*LinkedSlice, error) {
	lnk, err := linkerimpl.NewAlbums()
	if err != nil {
		return nil, err
	}
	var remName shared.RemoteName
	remName.FromString(remoteName)
	links, err := lnk.Links(remName)
	return newLinkedSlice(links), err
}

// Artists links.
func LinkerLinksArtists(remoteName string) (*LinkedSlice, error) {
	lnk, err := linkerimpl.NewArtists()
	if err != nil {
		return nil, err
	}
	var remName shared.RemoteName
	remName.FromString(remoteName)
	links, err := lnk.Links(remName)
	return newLinkedSlice(links), err
}

// Tracks links.
func LinkerLinksTracks(remoteName string) (*LinkedSlice, error) {
	lnk, err := linkerimpl.NewTracks()
	if err != nil {
		return nil, err
	}
	var remName shared.RemoteName
	remName.FromString(remoteName)
	links, err := lnk.Links(remName)
	return newLinkedSlice(links), err
}

// Get count of not matched albums.
func LinkerNotMatchedAlbumsCount(remoteName string) (int, error) {
	var remName shared.RemoteName
	remName.FromString(remoteName)
	lnk, err := linkerimpl.NewAlbums()
	if err != nil {
		return 9, err
	}
	return lnk.NotMatchedCount(remName)
}

// Get count of not matched artists.
func LinkerNotMatchedArtistsCount(remoteName string) (int, error) {
	var remName shared.RemoteName
	remName.FromString(remoteName)
	lnk, err := linkerimpl.NewArtists()
	if err != nil {
		return 9, err
	}
	return lnk.NotMatchedCount(remName)
}

// Get count of not matched tracks.
func LinkerNotMatchedTracksCount(remoteName string) (int, error) {
	var remName shared.RemoteName
	remName.FromString(remoteName)
	lnk, err := linkerimpl.NewTracks()
	if err != nil {
		return 9, err
	}
	return lnk.NotMatchedCount(remName)
}

// Import linker links.
func LinkerImportAlbumsLinks(from IoReader) error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewAlbums()
		if err != nil {
			return err
		}
		return lnk.Import(from)
	})
}

// Import linker links.
func LinkerImportArtistsLinks(from IoReader) error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewArtists()
		if err != nil {
			return err
		}
		return lnk.Import(from)
	})
}

// Import linker links.
func LinkerImportTracksLinks(from IoReader) error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewTracks()
		if err != nil {
			return err
		}
		return lnk.Import(from)
	})
}

// Export linker links.
func LinkerExportAlbumsLinks(dest IoWriter) error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewAlbums()
		if err != nil {
			return err
		}
		return lnk.Export(dest)
	})
}

// Export linker links.
func LinkerExportArtistsLinks(dest IoWriter) error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewArtists()
		if err != nil {
			return err
		}
		return lnk.Export(dest)
	})
}

// Export linker links.
func LinkerExportTracksLinks(dest IoWriter) error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewTracks()
		if err != nil {
			return err
		}
		return lnk.Export(dest)
	})
}

// Delete entities.
func LinkerCleanAlbums() error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewAlbums()
		if err != nil {
			return err
		}
		return lnk.Clean()
	})
}

// Delete entities.
func LinkerCleanArtists() error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewArtists()
		if err != nil {
			return err
		}
		return lnk.Clean()
	})
}

// Delete entities.
func LinkerCleanTracks() error {
	return execTask(0, func(ctx context.Context) error {
		lnk, err := linkerimpl.NewTracks()
		if err != nil {
			return err
		}
		return lnk.Clean()
	})
}

// Delete entities.
func LinkerCleanPlaylists() error {
	return execTask(0, func(ctx context.Context) error {
		repo := linkerimpl.NewPlaylistsRepository()
		return repo.DeleteAll()
	})
}

func newLinked(from linker.Linked) *Linked {
	return &Linked{
		self: from,
	}
}

type Linked struct {
	self linker.Linked
}

// Get entity id.
func (e Linked) EntityID() string {
	return e.self.EntityID().String()
}

// If missing - returns empty string.
func (e Linked) RemoteID() string {
	remID := e.self.RemoteID()
	if remID == nil {
		return ""
	}
	return remID.String()
}

// Set remote ID (aka relink).
// If empty - marks as missing.
func (e Linked) SetRemoteID(id string) error {
	return execTask(0, func(ctx context.Context) error {
		remID := stringToRemoteIDPtr(id)
		return e.self.SetRemoteID(remID)
	})
}

func newLinkedSlice(links []linker.Linked) *LinkedSlice {
	converted := make([]Linked, len(links))
	for i := range converted {
		converted[i] = *newLinked(links[i])
	}
	return &LinkedSlice{
		wrap: &wrappedSlice[Linked]{items: converted},
	}
}

type LinkedSlice struct {
	wrap *wrappedSlice[Linked]
}

func (e *LinkedSlice) Item(i int) *Linked {
	return e.wrap.Item(i)
}

func (e *LinkedSlice) Len() int {
	return e.wrap.Len()
}
