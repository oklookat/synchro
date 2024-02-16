package config

import (
	"errors"
	"net/url"
)

const _spotify = "spotify"

type Spotify struct {
	Host string `json:"host" mapstructure:"host"`
	Port int    `json:"port" mapstructure:"port"`
}

func (c Spotify) Key() string {
	return _spotify
}

func (c *Spotify) Default() {
	c.Host = "http://localhost"
	c.Port = 8080
}

func (c Spotify) Validate() error {
	if _, err := url.Parse(c.Host); err != nil {
		return err
	}
	if c.Port < 1023 {
		return errors.New("min port: 1024, max: 65535")
	}
	return nil
}
