package config

type VKMusic struct {
	BaseRemote
}

func (c *VKMusic) Default() {
}

func (c VKMusic) Validate() error {
	return nil
}
