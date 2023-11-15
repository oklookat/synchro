package deezer

import "github.com/oklookat/synchro/config"

type Config struct {
	Host string
	Port int
}

func (c Config) Key() config.Key {
	return "deezer"
}

func (c Config) Default() any {
	return Config{
		Host: "http://localhost",
		Port: 8081,
	}
}
