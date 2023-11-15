package linker

import "github.com/oklookat/synchro/config"

type Config struct {
	// Recheck missing entities?
	RecheckMissing bool
}

func (c Config) Key() config.Key {
	return "linker"
}

func (c Config) Default() any {
	return Config{
		RecheckMissing: true,
	}
}
