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
	viper.SetConfigFile(configPath)
	for _, cfg := range _configs {
		viper.SetDefault(cfg.Key().String(), cfg.Default())
	}

	err := viper.ReadInConfig()
	if err == nil {
		return err
	}

	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(configPath)
		if err != nil {
			return err
		}
		f.Close()
		if err = viper.WriteConfig(); err != nil {
			return err
		}
	}

	return viper.ReadInConfig()
}

// Get config by key.
func Get(key Key, toPtr any) error {
	return viper.UnmarshalKey(key.String(), toPtr)
}

// Set config.
func Set(key Key, src any) error {
	viper.Set(key.String(), src)
	return viper.WriteConfig()
}
