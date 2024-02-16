package commander

import (
	"errors"

	"github.com/gookit/event"
	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/darius"
	"github.com/oklookat/synchro/linking/linker"
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
	"github.com/oklookat/synchro/syncing/syncerimpl"

	// For go mod tidy (without it removes dependency).
	_ "golang.org/x/mobile/bind"
)

const (
	_packageName = "commander"
	_version     = "synchro-1.0.0"
)

type (
	OnUrler interface {
		OnURL(string)
	}

	OnUrlCoder interface {
		OnUrlCode(url, code string)
	}

	// https://pkg.go.dev/io#Reader
	IoReader interface {
		Read(p []byte) (n int, err error)
	}

	// https://pkg.go.dev/io#Writer
	IoWriter interface {
		Write(p []byte) (n int, err error)
	}

	OnLogger interface {
		OnLog(level int, msg string)
	}
)

var (
	ErrAnotherTaskInProgess = errors.New("another task in progress")
	_isBooted               = false
	ErrBootedBefore         = errors.New("booted before")
	_log                    logger.Logger
	_remotes                = map[shared.RemoteName]shared.Remote{
		yandexmusic.RemoteName: &yandexmusic.Remote{},
		spotify.RemoteName:     &spotify.Remote{},
		zvuk.RemoteName:        &zvuk.Remote{},
		vkmusic.RemoteName:     &vkmusic.Remote{},
		deezer.RemoteName:      &deezer.Remote{},
	}
)

func NewConfigGeneral() *ConfigGeneral {
	cfg := &config.General{}
	if err := config.Get(cfg); err != nil {
		cfg.Default()
	}
	return &ConfigGeneral{
		self: cfg,
	}
}

type ConfigGeneral struct {
	self *config.General
}

func (e *ConfigGeneral) SetDebug(val bool) error {
	e.self.Debug = val
	return config.Save(e.self)
}

func (e *ConfigGeneral) Debug() bool {
	return e.self.Debug
}

func Version() string {
	return _version
}

func Boot(dataPath string) error {
	if _isBooted {
		return ErrBootedBefore
	}

	// Logger (dummy).
	_log = logger.Get()

	if err := darius.Boot(dataPath); err != nil {
		return err
	}
	if err := config.Boot(); err != nil {
		return err
	}

	// Logger (from config).
	generalCfg := &config.General{}
	if err := config.Get(generalCfg); err != nil {
		generalCfg.Default()
	}
	logger.Boot(generalCfg.Debug)
	_log = logger.Get()
	config.SetLogger()

	// Repo.
	if err := repository.Boot(_remotes); err != nil {
		return err
	}

	// Core.
	linker.Boot()
	linkerimpl.Boot(_remotes)
	syncerimpl.Boot()
	lisybridge.Boot(_remotes)

	_log.Info("✨ Welcome to synchro!")
	_log.Info("🔗 https://github.com/oklookat/synchro")
	_log.Info("🔗 https://donationalerts.com/r/oklookat")
	_log.Info("🔗 https://boosty.to/oklookat/donate")

	_isBooted = true
	return nil
}

var (
	_isLoggerInit bool
)

func SetOnLogger(l OnLogger) error {
	if shared.IsNil(l) {
		return errors.New("SetOnLogger: nil OnLogger")
	}
	if _isLoggerInit {
		return errors.New("SetOnLogger: allowed only once")
	}
	_isLoggerInit = true
	event.On(string(shared.OnLog), event.ListenerFunc(func(e event.Event) error {
		level := e.Data()[shared.FieldLevel].(int)
		msg := e.Data()[shared.FieldMsg].(string)
		l.OnLog(level, msg)
		return nil
	}), event.Normal)
	return nil
}
