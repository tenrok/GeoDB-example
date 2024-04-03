package loggerx

import (
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*lumberjack.Logger
	mu      sync.RWMutex
	enabled bool
}

// New
func New(logger *lumberjack.Logger, enabled bool) *Logger {
	l := new(Logger)
	l.Logger = logger
	l.enabled = enabled
	return l
}

// Write implements io.Writer
func (l *Logger) Write(p []byte) (int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.enabled {
		return l.Logger.Write(p)
	}
	return len(p), nil
}

// Close implements io.Closer
func (l *Logger) Close() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.enabled {
		return l.Logger.Close()
	}
	return nil
}

// IsEnabled
func (l *Logger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled
}

// SetEnabled
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}
