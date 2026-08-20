package log

import (
	"mockservice/backend/common"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Init initializes the global logger instance
// Call this once at application startup
func Init(cfg *common.StarterConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(cfg)
	})
	return err
}

// Get returns the global logger instance
// Returns a no-op logger if not initialized
func Get() *zap.Logger {
	if globalLogger == nil {
		return zap.NewNop()
	}
	return globalLogger
}

// NewLogger creates a new logger instance
func NewLogger(cfg *common.StarterConfig) (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30,   // days
		Compress:   true, // disabled by default
	}

	var writer zapcore.WriteSyncer
	if cfg.Image {
		writer = zapcore.AddSync(os.Stdout)
	}

	if !cfg.Image && cfg.Debug {
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberjackLogger), zapcore.AddSync(os.Stdout))
	} else {
		writer = zapcore.AddSync(lumberjackLogger)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writer,
		zap.InfoLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger, nil
}
