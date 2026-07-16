// Package logger is a thin wrapper over log/slog that emits structured JSON
// logs. The package-level functions keep a small, stable API so call sites stay
// unchanged; the handler adds the source location to every record.
package logger

import (
	"fmt"
	"log/slog"
	"os"
)

var log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level:     slog.LevelInfo,
	AddSource: true,
}))

func Debug(msg ...interface{}) { log.Debug(fmt.Sprint(msg...)) }

func Debugf(format string, args ...interface{}) { log.Debug(fmt.Sprintf(format, args...)) }

func Info(msg ...interface{}) { log.Info(fmt.Sprint(msg...)) }

func Infof(format string, args ...interface{}) { log.Info(fmt.Sprintf(format, args...)) }

func Error(msg ...interface{}) { log.Error(fmt.Sprint(msg...)) }

func Errorf(format string, args ...interface{}) { log.Error(fmt.Sprintf(format, args...)) }

// L returns the underlying structured logger for call sites that log key/value
// attributes (for example the gRPC access log).
func L() *slog.Logger { return log }
