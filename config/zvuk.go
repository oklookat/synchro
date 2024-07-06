package config

type Zvuk struct {
	BaseRemote
}

func (c *Zvuk) Default() {
}

func (c Zvuk) Validate() error {
	return nil
}
