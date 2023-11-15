package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

type Key string

const (
	KeyGeneral Key = "general"
)

func (c Key) String() string {
	return string(c)
}

var (
	_configs = []Configer{
		General{},
	}
)

type Configer interface {
	// Config key.
	Key() Key
	// Struct with default values.
	Default() any
}

func Add(cfg Configer) {
	_configs = append(_configs, cfg)
}

func Boot(configPath string) error {
	for _, cfg := range _configs {
		viper.SetDefault(cfg.Key().String(), cfg.Default())
	}

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			f, err := os.Create(configPath)
			if err != nil {
				return err
			}
			f.Close()
			if err = viper.SafeWriteConfig(); err != nil {
				return err
			}
			return errors.New("Config file created. You must set you own config values. Config path: " + configPath)
		} else {
			return err
		}
	}

	return nil
}

// Get config by key.
func Get(key Key, toPtr any) error {
	return viper.UnmarshalKey(key.String(), toPtr)
}
