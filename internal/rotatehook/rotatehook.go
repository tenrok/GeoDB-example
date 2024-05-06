package rotatehook

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	LocalTime  bool
	Compress   bool
	Formatter  logrus.Formatter
	Level      logrus.Level
	Enabled    bool
}

type RotateHook struct {
	logger *lumberjack.Logger
	mu     sync.RWMutex
	cfg    *Config
}

func NewRotateHook(cfg *Config) *RotateHook {
	h := &RotateHook{
		cfg: cfg,
		logger: &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			LocalTime:  cfg.LocalTime,
			Compress:   cfg.Compress,
		},
	}

	return h
}

func (h *RotateHook) Levels() []logrus.Level {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return logrus.AllLevels[:h.cfg.Level+1]
}

func (h *RotateHook) Fire(entry *logrus.Entry) (err error) {
	b, err := h.cfg.Formatter.Format(entry)
	if err != nil {
		return err
	}
	if h.Enabled() {
		h.logger.Write(b)
	}

	return nil
}

func (h *RotateHook) Enabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cfg.Enabled
}

func (h *RotateHook) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cfg.Enabled = enabled
}
