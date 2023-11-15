package spotify

import "github.com/oklookat/synchro/config"

type Config struct {
	Host string
	Port int
}

func (c Config) Key() config.Key {
	return "spotify"
}

func (c Config) Default() any {
	return Config{
		Host: "http://localhost",
		Port: 8080,
	}
}
