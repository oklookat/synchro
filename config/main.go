package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Key string

const (
	KeyGeneral   Key = "general"
	KeyLinker    Key = "linker"
	KeySnapshots Key = "snapshots"
	KeySpotify   Key = "spotify"
	KeyDeezer    Key = "deezer"
)

var (
	_cfgFile *os.File
	_configs = map[Key]Configer{
		KeyGeneral:   &General{},
		KeyLinker:    &Linker{},
		KeySnapshots: &Snapshots{},
		KeySpotify:   &Spotify{},
		KeyDeezer:    &Deezer{},
	}
)

const (
	_fileFlags = os.O_CREATE | os.O_RDWR
	_filePerm  = 0666
)

type Configer interface {
	// Set default values.
	Default()
	// Validate.
	Validate() error
}

func Boot() error {
	for i := range _configs {
		_configs[i].Default()
	}

	cfgFile, err := os.OpenFile("config.json", _fileFlags, _filePerm)
	if err != nil {
		return err
	}
	_cfgFile = cfgFile

	if err := json.NewDecoder(_cfgFile).Decode(&_configs); err != nil {
		if err = Save(); err != nil {
			return err
		}
	}

	return validateSetAll()
}

func Save() error {
	if err := _cfgFile.Close(); err != nil {
		return err
	}
	err := os.Truncate(_cfgFile.Name(), 0)
	if err != nil {
		return err
	}
	if _cfgFile, err = os.OpenFile(_cfgFile.Name(), _fileFlags, _filePerm); err != nil {
		return err
	}
	enc := json.NewEncoder(_cfgFile)
	enc.SetIndent("", "\t")
	return enc.Encode(&_configs)
}

func Get[T any](what Key) (*T, error) {
	cfg, ok := _configs[what]
	if !ok {
		return new(T), errors.New("unknown key: " + string(what))
	}
	cfgTyped, _ := cfg.(T)
	return &cfgTyped, nil
}

func validateSetAll() error {
	for _, cfg := range _configs {
		if err := cfg.Validate(); err != nil {
			cfg.Default()
		}
	}
	return Save()
}
