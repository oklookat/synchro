package config

import (
	"errors"
)

type Snapshots struct {
	// Try to restore account library if account sync failed?
	AutoRecover bool `json:"autoRecover"`

	// Create snapshot when sync session starts?
	CreateWhenSyncing bool `json:"createWhenSyncing"`

	// Max auto snapshots count.
	MaxAuto int `json:"maxAuto"`
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
