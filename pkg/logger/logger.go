package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(logfile string) error {
	if logfile == "" {
		logfile = "./internal/log/app.log"
	}

	dir := "./internal/log"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	logLevelStr := os.Getenv("LOG_LEVEL")
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevelStr)); err != nil {
		level = zap.InfoLevel
	}

	logFile, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoderConfig.TimeKey = "time"
	consoleEncoderConfig.LevelKey = "level"
	consoleEncoderConfig.NameKey = "logger"
	consoleEncoderConfig.CallerKey = "caller"
	consoleEncoderConfig.MessageKey = "msg"
	consoleEncoderConfig.StacktraceKey = "stacktrace"

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)

	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.TimeKey = "time"
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	fileWriter := zapcore.AddSync(logFile)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)

	core := zapcore.NewTee(fileCore, consoleCore)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}