package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewLogger_Success(t *testing.T) {
	l, err := NewLogger("info")
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	l, err := NewLogger("invalid-level")
	assert.Error(t, err)
	assert.Nil(t, l)
	assert.Contains(t, err.Error(), "не удалось разобрать уровень логирования")
}

func TestLogger_LogInfo(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	l := &logger{logger: testLogger}

	l.LogInfo("test message", errors.New("test error"))
}

func TestLogger_LogStringInfo(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	l := &logger{logger: testLogger}

	l.LogStringInfo("test message", "key", "val")
}
