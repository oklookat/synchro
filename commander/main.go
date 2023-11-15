package commander

import (
	"os"
	"path"

	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/linking/linker"
	"github.com/oklookat/synchro/linking/linkerimpl"
	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/repository"
	"github.com/oklookat/synchro/streaming"
	"github.com/oklookat/synchro/streamings/deezer"
	"github.com/oklookat/synchro/streamings/spotify"
	"github.com/oklookat/synchro/streamings/vkmusic"
	"github.com/oklookat/synchro/streamings/yandexmusic"
	"github.com/oklookat/synchro/streamings/zvuk"
)

const (
	_dataDir = "./data"
)

var (
	_log     *logger.Logger
	_remotes = map[streaming.ServiceName]streaming.Service{
		yandexmusic.ServiceName: &yandexmusic.Service{},
		spotify.ServiceName:     &spotify.Service{},
		zvuk.ServiceName:        &zvuk.Service{},
		vkmusic.ServiceName:     &vkmusic.Service{},
		deezer.ServiceName:      &deezer.Service{},
	}
	_configPath = path.Join(_dataDir, "config.yml")
	_logPath    = path.Join(_dataDir, "log")
	_dbPath     = path.Join(_dataDir, "data.sqlite")
)

func Boot() error {
	// Data dir.
	if err := os.MkdirAll(_dataDir, 0644); err != nil {
		return err
	}

	// Config.
	config.Add(logger.Config{})
	config.Add(spotify.Config{})
	config.Add(deezer.Config{})
	config.Add(linker.Config{})
	if err := config.Boot(_configPath); err != nil {
		return err
	}

	// Logger.
	if err := os.MkdirAll(_logPath, 0644); err != nil {
		return err
	}
	loggerCfg := &logger.Config{}
	if err := config.Get(logger.ConfigKey, loggerCfg); err != nil {
		panic(err)
	}
	_log = logger.Boot(_logPath, *loggerCfg)

	// Repository.
	if err := repository.Boot(_dbPath, _log.With().Str("package", "repository").Logger(), _remotes); err != nil {
		return err
	}

	// Core.
	linker.Boot()
	linkerimpl.Boot(_remotes)

	_log.Info().Msg("✨ Welcome to synchro!")
	_log.Info().Msg("💽 https://github.com/oklookat/synchro")
	_log.Info().Msg("💵 https://donationalerts.com/r/oklookat")
	_log.Info().Msg("💵 https://boosty.to/oklookat/donate")
}
