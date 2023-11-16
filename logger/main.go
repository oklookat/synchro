package logger

import (
	"io"
	"os"
	"path"

	"github.com/oklookat/synchro/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const ConfigKey config.Key = "logger"

// Configuration for logging.
type Config struct {
	// Enable console logging.
	Console bool

	// Makes the log framework log JSON.
	AsJson bool

	// Log to a file.
	// the fields below can be skipped if this value is false!
	File bool

	// MaxSize the max size in MB of the logfile before it's rolled.
	MaxSize int

	// MaxBackups the max number of rolled files to keep.
	MaxBackups int

	// MaxAge the max age in days to keep a logfile.
	MaxAge int

	// Logging level.
	Level int8
}

func (c Config) Key() config.Key {
	return ConfigKey
}

func (c Config) Default() any {
	return Config{
		Console:    true,
		AsJson:     false,
		File:       true,
		MaxSize:    5,
		MaxBackups: 5,
		MaxAge:     5,
		Level:      int8(zerolog.TraceLevel),
	}
}

type Logger struct {
	*zerolog.Logger
}

func Boot(logsDir string, config Config) (*Logger, error) {
	var writers []io.Writer

	if config.Console {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if config.File {
		rf, err := newRollingFile(logsDir, config)
		if err != nil {
			return nil, err
		}
		writers = append(writers, rf)
	}
	mw := io.MultiWriter(writers...)

	zerolog.SetGlobalLevel(zerolog.Level(config.Level))
	logger := zerolog.New(mw).With().Timestamp().Logger()

	return &Logger{
		Logger: &logger,
	}, nil
}

func newRollingFile(logsDir string, config Config) (io.Writer, error) {
	if err := os.MkdirAll(logsDir, 0744); err != nil {
		return nil, err
	}

	return &lumberjack.Logger{
		Filename:   path.Join(logsDir, "log.log"),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}, nil
}
