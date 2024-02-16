package shared

type Event string

func (e Event) String() string {
	return string(e)
}

const (
	FieldMsg   = "msg"
	FieldLevel = "level"
)

const (
	// No fields.
	OnAutoSyncStart Event = "OnAutoSyncStart"
	OnConfigChanged Event = "OnConfigChanged"

	// "Msg", "Level" fields.
	OnLog Event = "OnLog"
)
