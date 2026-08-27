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

	outputs := logCfg.Outputs
	if len(outputs) == 0 {
		outputs = []string{"console"}
	}

	cores := make([]zapcore.Core, 0, len(outputs))
	for _, output := range outputs {
		switch strings.ToLower(strings.TrimSpace(output)) {
		case "console":
			cores = append(cores, zapcore.NewCore(newConsoleEncoder(dev), zapcore.AddSync(os.Stdout), level))
		case "file":
			file, err := openLogFile(logCfg)
			if err != nil {
				return nil, err
			}
			cores = append(cores, zapcore.NewCore(newFileEncoder(), zapcore.AddSync(file), level))
		}
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no valid logging outputs configured")
	}

	opts := []zap.Option{zap.AddStacktrace(zapcore.DPanicLevel)}
	if !dev {
		opts = append(opts, zap.AddCaller())
	}

	return zap.New(zapcore.NewTee(cores...), opts...), nil
}

func newConsoleEncoder(dev bool) zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        zapcore.OmitKey,
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if dev {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(cfg)
	}
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(cfg)
}

func newFileEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func openLogFile(logCfg config.LoggingConfig) (*os.File, error) {
	if err := os.MkdirAll(logCfg.Directory, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	fileName := fmt.Sprintf("%s-%s.log", logCfg.Level, time.Now().Format("2006-01-02"))
	path := filepath.Join(logCfg.Directory, fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return file, nil
}
