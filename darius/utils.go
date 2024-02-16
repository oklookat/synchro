package darius

import (
	"errors"
	"fmt"
	"os"
)

// Dir/file exists?
//
// Exists: returns nil.
//
// Not exists: returns wrapped os.ErrNotExist error.
//
// ???: returns os.Stat error.
func isExists(abs string) error {
	var err error
	if _, err = os.Stat(abs); err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(_errPrefix+`"%s" not exists. (%w)`, abs, os.ErrNotExist)
	}
	return fmt.Errorf(_errPrefix+`%w`, err)
}
