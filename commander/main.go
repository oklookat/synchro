package commander

import (
	"log/slog"
	"os"
	"runtime"

	"github.com/oklookat/synchro/commander/cli"
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
	const dataPath = "data"

	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return err
	}

	if err := config.Boot(dataPath + "/config.json"); err != nil {
		return err
	}

	// Logger (from config).
	genCfg, err := config.Get[*config.General](config.KeyGeneral)
	if err != nil {
		return err
	}

	logger.Boot(dataPath+"/log.log", (*genCfg).Debug)

	// Repo.
	if err := repository.Boot(dataPath+"/data.sqlite", _remotes); err != nil {
		return err
	}

	// Core.
	linkerimpl.Boot(_remotes)

	slog.Info("✨ Welcome to synchro!")
	slog.Info("🔗 https://github.com/oklookat/synchro")
	slog.Info("🔗 https://donationalerts.com/r/oklookat")
	slog.Info("🔗 https://boosty.to/oklookat/donate")

	// UI.
	if runtime.GOOS != "android" && runtime.GOOS != "ios" {
		if err := cli.Boot(); err != nil {
			return err
		}
	} else {
		slog.Info("🤔 Nice phone bro!")
	}

	return nil
}
