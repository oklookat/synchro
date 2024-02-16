package config

const _general = "general"

type General struct {
	Debug bool `json:"debug" mapstructure:"debug"`
}

func (c *General) Default() {
	c.Debug = true
}

func (c General) Key() string {
	return _general
}

func (c General) Validate() error {
	return nil
}
