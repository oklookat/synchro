package syncerimpl

import (
	"github.com/oklookat/synchro/logger"
)

var (
	_log *logger.Logger
)

func Boot() {
	_log = logger.WithPackageName("syncerimpl")
}
