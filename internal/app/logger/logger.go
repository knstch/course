package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Log struct {
	logger *zap.Logger
}

func InitLogger(fileName string) (*Log, error) {
	logger := Log{}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(&lumberjack.Logger{
			Filename:   `./log/` + fileName + `_logfile.log`,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		}), zap.InfoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(&lumberjack.Logger{
			Filename:   `./log/` + fileName + `_error.log`,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		}), zap.ErrorLevel),
	)

	logger.logger = zap.New(core)

	return &logger, nil
}

func (l *Log) Error(method, errMessage, message string, code int) {
	fields := []zapcore.Field{
		zap.String("method", method),
		zap.String("message", errMessage),
		zap.String("code", fmt.Sprint(code)),
	}

	l.logger.Error(message, fields...)
}

func (l *Log) Info(message, actionName, request string) {
	fields := []zapcore.Field{
		zap.String("method", actionName),
		zap.String("request", request),
	}

	l.logger.Info(message, fields...)
}
