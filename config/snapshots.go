package config

import (
	"errors"
)

const _snapshots = "snapshots"

type Snapshots struct {
	// Try to restore account library if account sync failed?
	AutoRecover bool `json:"autoRecover" mapstructure:"autoRecover"`

	// Create snapshot when sync session starts?
	CreateWhenSyncing bool `json:"createWhenSyncing" mapstructure:"createWhenSyncing"`

	// Max auto snapshots count.
	MaxAuto int `json:"maxAuto" mapstructure:"maxAuto"`
}

func (c Snapshots) Key() string {
	return _snapshots
}

func (c *Snapshots) Default() {
	c.AutoRecover = true
	c.CreateWhenSyncing = true
	c.MaxAuto = 30
}

func (c Snapshots) Validate() error {
	if c.MaxAuto < 2 {
		return errors.New("min: 2")
	}
	return nil
}
