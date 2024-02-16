package syncerimpl

import (
	"context"

	"github.com/oklookat/synchro/shared"
)

func NewMetaString(
	entityID shared.EntityID,
	data string,
	setData func(context.Context, string) error,
	syncSetting shared.SynchronizationSettings,
) MetaString {
	return MetaString{
		entityID:    entityID,
		data:        data,
		setData:     setData,
		syncSetting: syncSetting,
	}
}

func NewMetaBool(
	entityID shared.EntityID,
	data bool,
	setData func(context.Context, bool) error,
	syncSetting shared.SynchronizationSettings,
) MetaBool {
	return MetaBool{
		entityID:    entityID,
		data:        data,
		setData:     setData,
		syncSetting: syncSetting,
	}
}

// Implements MetaThing.
type MetaString struct {
	entityID    shared.EntityID
	data        string
	setData     func(context.Context, string) error
	syncSetting shared.SynchronizationSettings
}

func (e MetaString) EntityID() shared.EntityID {
	return e.entityID
}

func (e MetaString) Param() string {
	return e.data
}

func (e MetaString) SetParam(ctx context.Context, val string) error {
	return e.setData(ctx, val)
}

func (e MetaString) SyncSetting() shared.SynchronizationSettings {
	return e.syncSetting
}

// Implements MetaThing.
type MetaBool struct {
	entityID    shared.EntityID
	data        bool
	setData     func(context.Context, bool) error
	syncSetting shared.SynchronizationSettings
}

func (e MetaBool) EntityID() shared.EntityID {
	return e.entityID
}

func (e MetaBool) Param() bool {
	return e.data
}

func (e MetaBool) SetParam(ctx context.Context, val bool) error {
	return e.setData(ctx, val)
}

func (e MetaBool) SyncSetting() shared.SynchronizationSettings {
	return e.syncSetting
}
