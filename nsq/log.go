package nsq

import (
	"strings"

	nsq "github.com/nsqio/go-nsq"
)

// Logger should implement basic log output functions for all levels used in NSQ
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
}

// LogLevel specifies Producer or Consumer logging verbosity
type LogLevel int

func (ll LogLevel) toNSQ() nsq.LogLevel {
	switch ll {
	case LogLevelWarning:
		return nsq.LogLevelWarning
	case LogLevelInfo:
		return nsq.LogLevelInfo
	case LogLevelDebug:
		return nsq.LogLevelDebug
	}
	return nsq.LogLevelError
}

// Available log levels
const (
	LogLevelError LogLevel = iota
	LogLevelWarning
	LogLevelInfo
	LogLevelDebug
)

type logger struct {
	KeepNsqLookupD404 bool
	Logger
}

// Output implements stdlib log.Logger.Output using the underlying logger
func (l logger) Output(_ int, s string) error {
	s = strings.TrimSpace(s)
	var level string
	if len(s) >= 3 {
		level, s = s[:3], s[3:]
	}
	switch level {
	case nsq.LogLevelError.String():
		if !l.KeepNsqLookupD404 && strings.Contains(s, "TOPIC_NOT_FOUND") {
			return nil
		}
		l.Logger.Error(s)
	case nsq.LogLevelWarning.String():
		l.Logger.Warn(s)
	case nsq.LogLevelDebug.String():
		l.Logger.Debug(s)
	default:
		l.Logger.Info(s)
	}
	return nil
}
