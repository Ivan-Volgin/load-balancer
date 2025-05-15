package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const tsKey = "timestamp"

// NewLogger — создаёт настраиваемый логгер на основе библиотеки zap, с поддержкой уровня логирования, JSON-форматом и выводом в stdout.
// Возвращает SugaredLogger для удобства использования или ошибку, если не удалось создать логгер.
func NewLogger(level string) (*zap.SugaredLogger, error) {
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing log level: %s", level)
	}

	logger, err := zap.Config{
		Level:       logLevel,
		Encoding:    "json",
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
			TimeKey:    tsKey,
			EncodeTime: zapcore.RFC3339NanoTimeEncoder,
		},
		DisableStacktrace: true,
	}.Build()
	if err != nil {
		return nil, errors.Wrap(err, "Error building logger")
	}

	return logger.Sugar(), nil
}
