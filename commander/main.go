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
	_log        *logger.Logger
	_configPath = path.Join(_dataDir, "config.yml")
	_logsDir    = path.Join(_dataDir, "logs")
	_dbPath     = path.Join(_dataDir, "data.sqlite")
	_services   = map[streaming.ServiceName]streaming.Service{
		yandexmusic.ServiceName: &yandexmusic.Service{},
		spotify.ServiceName:     &spotify.Service{},
		zvuk.ServiceName:        &zvuk.Service{},
		vkmusic.ServiceName:     &vkmusic.Service{},
		deezer.ServiceName:      &deezer.Service{},
	}
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
	loggerCfg := &logger.Config{}
	if err := config.Get(loggerCfg.Key(), loggerCfg); err != nil {
		panic(err)
	}
	var err error
	if _log, err = logger.Boot(_logsDir, *loggerCfg); err != nil {
		panic(err)
	}

	// Repository.
	if err := repository.Boot(_dbPath, _log, _services); err != nil {
		return err
	}

	// Core.
	linker.Boot(_log)
	linkerimpl.Boot(_services)

	_log.Info().Msg("✨ Welcome to synchro!")
	_log.Info().Msg("💽 https://github.com/oklookat/synchro")
	_log.Info().Msg("💵 https://donationalerts.com/r/oklookat")
	_log.Info().Msg("💵 https://boosty.to/oklookat/donate")

	return nil
}
