package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type logger struct {
	logger *zap.Logger
}

// NewLogger инициализация логгера.
func NewLogger(level string) (*logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать уровень логирования: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("не удалось построить конфигурацию логгера: %w", err)
	}

	return &logger{logger: zl}, nil
}

// LogInfo вызов лога ошибок уровня Info.
func (l *logger) LogInfo(massage string, err error) {
	l.logger.Info(massage, zap.Error(err))
}

// CustomLogger интерфейс, который должен использоваться
// в других пакетах, где нужно логирование ошибок.
type CustomLogger interface {
	LogInfo(massage string, err error)
}

// LogStringInfo вызов лога key-value уровня Info.
func (l *logger) LogStringInfo(massage string, key, val string) {
	l.logger.Info(massage, zap.String(key, val))
}
