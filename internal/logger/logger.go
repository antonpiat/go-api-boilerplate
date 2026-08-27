package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func NewLogger(config *config.LoggingConfig) *zap.Logger {
	level := parseLevel(config.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "message",
		CallerKey:     "caller",
		FunctionKey:   "function",
		StacktraceKey: "stacktrace",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	cores := make([]zapcore.Core, len(config.Outputs))
	for i, output := range config.Outputs {
		switch output {
		case "console":
			cores[i] = zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		case "file":
			if err := os.MkdirAll(config.Directory, 0755); err != nil {
				panic(err)
			}

			fileName := fmt.Sprintf("%s-%s.log", config.Level, time.Now().Format("2006-01-02"))
			path := filepath.Join(config.Directory, fileName)

			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				panic(err)
			}
			cores[i] = zapcore.NewCore(fileEncoder, zapcore.AddSync(file), level)
		}
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}
