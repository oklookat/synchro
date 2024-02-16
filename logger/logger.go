package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type contextKey string

const (
	_ctxKey contextKey = "SYNCHRO_LOGGER"
)

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
)

type Level int

func (e Level) String() string {
	switch e {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelPanic:
		return "PANIC"
	}
	return "UNKNOWN"
}

func (e Level) Int() int {
	return int(e)
}

type Out func(level Level, data []byte) (int, error)

func New(out Out) Logger {
	if out == nil {
		out = func(level Level, data []byte) (int, error) {
			return os.Stderr.Write(data)
		}
	}
	return Logger{
		Out:          out,
		MinimalLevel: LevelDebug,
	}
}

// Get logger from context. Returns nil if not exits.
func FromContext(ctx context.Context) *Logger {
	what := ctx.Value(_ctxKey)
	log, ok := what.(*Logger)
	if !ok {
		return nil
	}
	return log
}

type Logger struct {
	Out          Out
	MinimalLevel Level

	fields map[string]any
}

// Copy current Logger to new.
func (e *Logger) copy() Logger {
	fields := map[string]any{}
	for k, v := range e.fields {
		fields[k] = v
	}
	return Logger{
		Out:          e.Out,
		MinimalLevel: e.MinimalLevel,
		fields:       fields,
	}
}

// Provide current logger pointer to context.
func (e *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxKey, e)
}

// Add field.
func (e Logger) AddField(name string, value any) Logger {
	cop := e.copy()
	if len(cop.fields) == 0 {
		cop.fields = map[string]any{}
	}
	cop.fields[name] = value
	return cop
}

func (e Logger) Trace(msg string) {
	e.log(LevelTrace, msg)
}

func (e Logger) Debug(msg string) {
	e.log(LevelDebug, msg)
}

func (e Logger) Info(msg string) {
	e.log(LevelInfo, msg)
}

func (e Logger) Warn(msg string) {
	e.log(LevelWarn, msg)
}

func (e Logger) Error(msg string) {
	e.log(LevelError, msg)
}

func (e Logger) Panic(msg string) {
	e.log(LevelPanic, msg)
}

func (e Logger) log(level Level, msg string) {
	if e.Out == nil || e.MinimalLevel > level {
		return
	}

	var fieldsFormatted []string
	for name, val := range e.fields {
		fieldsFormatted = append(fieldsFormatted, fmt.Sprintf("%s=%v;", name, val))
	}

	fullMsg := msg
	if len(fieldsFormatted) > 0 {
		fullMsg += " " + strings.Join(fieldsFormatted, " ")
	}

	e.Out(level, []byte(fullMsg))

	if level == LevelPanic {
		panic(fullMsg)
	}
}
