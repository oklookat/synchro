package linker

import (
	"github.com/oklookat/synchro/logger"
)

var (
	_log *logger.Logger
)

func Boot(log *logger.Logger) {
	_log = log
}
