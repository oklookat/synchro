package yandexmusic

import (
	"github.com/oklookat/synchro/config"
	"github.com/oklookat/synchro/shared"
)

type Config struct {
	DeviceID string
}

func (c Config) Key() config.Key {
	return "yandexMusic"
}

func (c Config) Default() any {
	return Config{
		DeviceID: shared.GenerateWord(12),
	}
}
