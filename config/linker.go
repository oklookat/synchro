package config

const _linker = "linker"

type Linker struct {
	// Recheck missing entities?
	RecheckMissing bool `json:"recheckMissing" mapstructure:"recheckMissing"`
}

func (c Linker) Key() string {
	return _linker
}

func (c *Linker) Default() {
	c.RecheckMissing = false
}

func (c Linker) Validate() error {
	return nil
}
