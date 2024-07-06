package config

import "github.com/oklookat/synchro/shared"

type YandexMusic struct {
	BaseRemote
	DeviceID string `json:"deviceID"`
}

func (c *YandexMusic) Default() {
	c.DeviceID = shared.GenerateULID()
}

func (c YandexMusic) Validate() error {
	return nil
}
