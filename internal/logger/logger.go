package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
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

func New(logCfg config.LoggingConfig, environment string) (*zap.Logger, error) {
	level := parseLevel(logCfg.Level)
	dev := strings.EqualFold(environment, "development") || strings.EqualFold(environment, "dev")

	baseEncoder := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	outputs := logCfg.Outputs
	if len(outputs) == 0 {
		outputs = []string{"console"}
	}

	cores := make([]zapcore.Core, 0, len(outputs))
	for _, output := range outputs {
		switch strings.ToLower(strings.TrimSpace(output)) {
		case "console":
			encCfg := baseEncoder
			var encoder zapcore.Encoder
			if dev {
				encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
				encoder = zapcore.NewConsoleEncoder(encCfg)
			} else {
				encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
				encoder = zapcore.NewJSONEncoder(encCfg)
			}
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
		case "file":
			if err := os.MkdirAll(logCfg.Directory, 0o755); err != nil {
				return nil, fmt.Errorf("create log directory: %w", err)
			}
			fileName := fmt.Sprintf("%s-%s.log", logCfg.Level, time.Now().Format("2006-01-02"))
			path := filepath.Join(logCfg.Directory, fileName)
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return nil, fmt.Errorf("open log file: %w", err)
			}
			encCfg := baseEncoder
			encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
			cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.AddSync(file), level))
		}
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no valid logging outputs configured")
	}

	core := zapcore.NewTee(cores...)
	log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return log, nil
}
