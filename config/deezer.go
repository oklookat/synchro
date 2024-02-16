package config

import (
	"errors"
	"net/url"
)

const _deezer = "deezer"

type Deezer struct {
	Host string `json:"host" mapstructure:"host"`
	Port int    `json:"port" mapstructure:"port"`
}

func (c Deezer) Key() string {
	return _deezer
}

func (c *Deezer) Default() {
	c.Host = "http://localhost"
	c.Port = 8081
}

func (c Deezer) Validate() error {
	if _, err := url.Parse(c.Host); err != nil {
		return err
	}
	if c.Port < 1023 {
		return errors.New("min port: 1024, max: 65535")
	}
	return nil
}
