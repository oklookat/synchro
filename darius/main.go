/*
abstraction over %APPDATA%/vendor/appname dir.
*/
package darius

const (
	_errPrefix = "darius: "
	_PERM      = 0644
)

var (
	_appData = appData{}
)

func Boot(dataPath string) error {
	if len(dataPath) > 0 {
		_appData = appData{abs: dataPath}
	}
	return _appData.Boot(dataPath, "oklookat", "synchro")
}

// Get absolute path to appdata.
func Abs() string {
	return _appData.abs
}

// Add file and create if not exists.
func WrapFile(fpath string) (*File, error) {
	file := &File{}
	if err := file.new(fpath); err != nil {
		return nil, err
	}

	return file, nil
}
