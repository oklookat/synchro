// App directory wrapper.
package darius

import (
	"errors"
	"os"
	"path"
	"path/filepath"
)

// Wrapper over "%APPDATA%/vendor/appname".
type appData struct {
	// absolute path to %APPDATA%/vendor/appname
	abs string
}

func (a *appData) Boot(dataPath, vendor, appName string) error {
	if len(dataPath) > 0 {
		a.abs = dataPath
		return nil
	}

	// get absolute path
	appDataPath, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	a.abs = path.Join(appDataPath, vendor, appName)

	// create dirs if not exists
	if err = isExists(a.abs); err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(a.abs, _PERM)
	}

	return err
}

// Relative path to absoulute.
//
// Output: %APPDATA%/vendor/appname/rel/path/etc
func (a appData) RelToAbs(rel string) (string, error) {
	abs := path.Join(a.abs, rel)
	abs = filepath.Clean(filepath.FromSlash(abs))
	return abs, nil
}
