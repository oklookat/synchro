package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Key string

const (
	KeyDeezer      Key = "deezer"
	KeyGeneral     Key = "general"
	KeyLinker      Key = "linker"
	KeySpotify     Key = "spotify"
	KeyVKMusic     Key = "vkMusic"
	KeyYandexMusic Key = "yandexMusic"
	KeyZvuk        Key = "zvuk"
)

var (
	_cfgFile *os.File
	_configs = map[Key]Configer{
		KeyDeezer:      &Deezer{},
		KeyGeneral:     &General{},
		KeyLinker:      &Linker{},
		KeySpotify:     &Spotify{},
		KeyVKMusic:     &VKMusic{},
		KeyYandexMusic: &YandexMusic{},
		KeyZvuk:        &Zvuk{},
	}
)

const (
	_fileFlags = os.O_CREATE | os.O_RDWR
	_filePerm  = 0666
)

type (
	Configer interface {
		// Set default values.
		Default()
		// Validate.
		Validate() error
	}

	Proxy struct {
		Proxy bool `json:"proxy"`
		// Example: http://proxyIp:proxyPort
		URL string `json:"url"`
	}

	BaseRemote struct {
		Proxy Proxy `json:"proxy"`
	}
)

func Boot(cfgPath string) error {
	// Set defaults.
	for i := range _configs {
		_configs[i].Default()
	}

	cfgFile, err := os.OpenFile(cfgPath, _fileFlags, _filePerm)
	if err != nil {
		return err
	}
	_cfgFile = cfgFile
	defer _cfgFile.Close()

	// Read and merge configuration.
	decoder := json.NewDecoder(_cfgFile)
	var rawConfigs map[Key]json.RawMessage
	if err := decoder.Decode(&rawConfigs); err != nil {
		if err = Save(); err != nil {
			return err
		}
	} else {
		for key, raw := range rawConfigs {
			if cfg, exists := _configs[key]; exists {
				if err := json.Unmarshal(raw, cfg); err != nil {
					cfg.Default()
				}
			}
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
		return nil, errors.New("unknown key: " + string(what))
	}
	cfgTyped, ok := cfg.(T)
	if !ok {
		return nil, errors.New("config assertion not ok. Wtf")
	}
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
