package config

type Linker struct {
	// Recheck missing entities?
	RecheckMissing bool `json:"recheckMissing"`
}

func (c *Linker) Default() {
	c.RecheckMissing = false
}

func (c Linker) Validate() error {
	return nil
}
