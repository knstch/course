// logger содержит методы логирования.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Error(message, method, errMessage string, code int)
	Info(message, method, request string)
}

// Log хранит внутри себя сущность zap.Logger, которая содержит методы логирования.
type Log struct {
	logger *zap.Logger
}

// InitLogger инициализирует логгер, в качестве параметра принимает названия файла лога.
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

// Error используется для создания лога с ошибкой, принимает в качестве параметров
// метод, ошибку, комментарий об ошибке, и код ошибки.
func (l *Log) Error(method, errMessage, message string, code int) {
	fields := []zapcore.Field{
		zap.String("method", method),
		zap.String("message", errMessage),
		zap.String("code", fmt.Sprint(code)),
	}

	l.logger.Error(message, fields...)
}

// Info используется для создания лога уровня Info, принимает в себя сообщение, метод и параметры запроса.
func (l *Log) Info(message, method, request string) {
	fields := []zapcore.Field{
		zap.String("method", method),
		zap.String("request", request),
	}

	l.logger.Info(message, fields...)
}
