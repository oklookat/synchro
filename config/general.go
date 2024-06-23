package config

type General struct {
	Debug bool `json:"debug"`
}

func (c *General) Default() {
	c.Debug = true
}

func (c General) Validate() error {
	return nil
}
