package logger

import (
	"log/slog"
	"os"
	"sync"
)

func Boot(
	debug bool,
) error {
	const (
		logPath      = "log.log"
		maxSizeBytes = 1000000
	)

	writer, err := newLogWriter(logPath, maxSizeBytes)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{}
	if debug {
		opts.Level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(writer, opts))

	slog.SetDefault(logger)
	return err
}

func newLogWriter(
	logPath string,
	maxLogSizeBytes int64,
) (*logWriter, error) {
	writer := &logWriter{
		maxLogSizeBytes: maxLogSizeBytes,
	}
	if err := writer.openLog(logPath); err != nil {
		return nil, err
	}
	logFileSize, err := writer.logFileSize()
	if err != nil {
		return nil, err
	}
	writer.fileSize = logFileSize

	return writer, err
}

type logWriter struct {
	logPath         string
	logMutex        sync.Mutex
	maxLogSizeBytes int64
	fileSize        int64
	file            *os.File
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)

	l.logMutex.Lock()
	if (l.fileSize + int64(len(p))) > l.maxLogSizeBytes {
		l.file.Close()
		if err = os.Truncate(l.file.Name(), 0); err != nil {
			return 0, err
		}
		l.fileSize = 0
		l.openLog(l.logPath)
	}
	n, err = l.file.Write(p)
	l.fileSize += int64(n)
	l.logMutex.Unlock()

	return n, err
}

func (l *logWriter) logFileSize() (int64, error) {
	fStat, err := os.Stat(l.file.Name())
	if err != nil {
		return 0, err
	}
	return fStat.Size(), err
}

func (l *logWriter) openLog(logPath string) error {
	if l.file != nil {
		l.file.Close()
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	l.file = logFile
	l.logPath = logPath
	return err
}
