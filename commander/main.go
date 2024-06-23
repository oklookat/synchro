package commander

import (
	"log/slog"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/remote/deezer"
	"github.com/oklookat/synchro/remote/spotify"
	"github.com/oklookat/synchro/remote/vkmusic"
	"github.com/oklookat/synchro/remote/yandexmusic"
	"github.com/oklookat/synchro/remote/zvuk"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/shared"
	"github.com/oklookat/synchro/syncing/lisybridge"
)

var (
	_remotes = map[shared.RemoteName]shared.Remote{
		yandexmusic.RemoteName: &yandexmusic.Remote{},
		spotify.RemoteName:     &spotify.Remote{},
		zvuk.RemoteName:        &zvuk.Remote{},
		vkmusic.RemoteName:     &vkmusic.Remote{},
		deezer.RemoteName:      &deezer.Remote{},
	}
)

func Boot() error {
	if err := config.Boot(); err != nil {
		return err
	}

	// Logger (from config).
	genCfg, err := config.Get[config.General](config.KeyGeneral)
	if err != nil {
		return err
	}

	logger.Boot(genCfg.Debug)

	// Repo.
	if err := repository.Boot(_remotes); err != nil {
		return err
	}

	// Core.
	linkerimpl.Boot(_remotes)
	lisybridge.Boot(_remotes)

	slog.Info("✨ Welcome to synchro!")
	slog.Info("🔗 https://github.com/oklookat/synchro")
	slog.Info("🔗 https://donationalerts.com/r/oklookat")
	slog.Info("🔗 https://boosty.to/oklookat/donate")

	return nil
}
