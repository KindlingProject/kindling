package logger

import (
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"testing"
)

func Test_newLogger(t *testing.T) {
	config := &lumberjack.Logger{
		Filename:   "tmp.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,
		LocalTime:  true,
		Compress:   false,
	}
	logger := CreateCombineLogger(config)
	logger.Info("This is a Info test", zap.String("param1", "value1"))
	logger.Debug("This is a Info test", zap.String("param1", "value1"))
}
