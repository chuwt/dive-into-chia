package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func NewLog(logLevel string, field ...zap.Field) *zap.Logger {
	if logger == nil {
		SetLogger(logLevel)
	}

	return logger.With(field...)
}

func SetLogger(logLevel string) {
	var err error

	logConf := zap.NewProductionConfig()
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder

	logConf.EncoderConfig = encoder
	logConf.Encoding = "console"

	var level zapcore.Level
	if err = level.Set(logLevel); err != nil {
		panic(err)
	}
	logConf.Level = zap.NewAtomicLevelAt(level)

	logger, err = logConf.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
