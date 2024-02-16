package config

import (
	"fmt"

	"github.com/gookit/event"
	"github.com/oklookat/synchro/darius"
	"github.com/oklookat/synchro/logger"
	"github.com/oklookat/synchro/shared"
	"github.com/spf13/viper"
)

var (
	_log        *logger.Logger
	_configFile *darius.File
)

var configs = []Configer{
	&General{},
	&Linker{},
	&Snapshots{},
	&Spotify{},
	&Deezer{},
}

func Boot() error {
	setDefaults()

	filed, err := darius.WrapFile("config.json")
	if err != nil {
		return err
	}
	_configFile = filed
	viper.SetConfigFile(filed.Abs())

	if err := viper.ReadInConfig(); err != nil {
		// Not found.
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create.
			if err = _configFile.CreateIfNotExists(); err != nil {
				return err
			}
			return viper.WriteConfig()
		}

		// Bad format.
		if _, ok := err.(viper.ConfigParseError); ok {
			// Clean, write.
			if err = _configFile.Clean(); err != nil {
				return err
			}
			return viper.WriteConfig()
		}

		return fmt.Errorf("config: %e", err)
	}

	return validateSetAll()
}

func SetLogger() {
	_log = logger.WithPackageName("config")
}

func Save(what Configer) error {
	if err := what.Validate(); err != nil {
		return err
	}
	// json and mapstructure tags are ignored (idk why).
	viper.Set(what.Key(), what)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	event.Fire(shared.OnConfigChanged.String(), nil)
	return nil
}

func Get(who Configer) error {
	var err error
	if err = viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.UnmarshalKey(who.Key(), who); err != nil {
		_log.AddField("key", who.Key()).
			Error("save: " + err.Error())
	}
	return err
}

func validateSetAll() error {
	for _, cfg := range configs {
		var err error
		if err = Get(cfg); err == nil {
			continue
		}
		cfg.Default()
		if err := Save(cfg); err != nil {
			_log.Error("save: " + err.Error())
			return err
		}
	}
	return nil
}

func setDefaults() {
	for _, cfg := range configs {
		cfg.Default()
		viper.SetDefault(cfg.Key(), cfg)
	}
}

type Configer interface {
	Key() string
	Default()
	Validate() error
}
