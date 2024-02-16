package logger

import (
	"github.com/gookit/event"
	"github.com/oklookat/synchro/shared"
)

var self = New(nil)

func Boot(debug bool) {
	if debug {
		self.MinimalLevel = LevelDebug
	} else {
		self.MinimalLevel = LevelInfo
	}

	self.Out = func(level Level, data []byte) (int, error) {
		if len(data) == 0 {
			return 0, nil
		}
		event.Fire(shared.OnLog.String(), map[string]interface{}{
			shared.FieldLevel: level.Int(),
			shared.FieldMsg:   string(data),
		})
		return len(data), nil
	}
}

func Get() Logger {
	return self
}

func WithPackageName(packageName string) *Logger {
	log := self.AddField("package", packageName)
	return &log
}
