package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aurora-go/aurora/internal/config"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

// InitZapLogger 初始化 Zap 日志系统
func InitZapLogger(cfg *config.LogConfig) {
	level := parseLogLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 输出到stdout
	consoleWriter := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleWriter, level),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Sugar = Logger.Sugar()

	slog.SetDefault(slog.New(&zapLogAdapter{logger: Sugar}))

	slog.Info("Zap logger initialized",
		"level", cfg.Level,
		"format", cfg.Format,
	)
}

// Sync 刷新缓冲区日志
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// GetLogger 获取原始 Logger 实例
func GetLogger() *zap.Logger {
	return Logger
}

// GetSugarLogger 获取 SugaredLogger 实例
func GetSugarLogger() *zap.SugaredLogger {
	return Sugar
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// zapLogAdapter 将 slog 桥接到 zap
type zapLogAdapter struct {
	logger *zap.SugaredLogger
}

func (a *zapLogAdapter) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

func (a *zapLogAdapter) Handle(_ context.Context, record slog.Record) error {
	args := make([]any, 0, record.NumAttrs()*2)
	record.Attrs(func(attr slog.Attr) bool {
		args = append(args, attr.Key, attr.Value.Any())
		return true
	})

	msg := record.Message
	switch record.Level {
	case slog.LevelDebug:
		a.logger.Debugw(msg, args...)
	case slog.LevelInfo:
		a.logger.Infow(msg, args...)
	case slog.LevelWarn:
		a.logger.Warnw(msg, args...)
	case slog.LevelError:
		a.logger.Errorw(msg, args...)
	default:
		a.logger.Infow(msg, args...)
	}
	return nil
}

func (a *zapLogAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return a
}

func (a *zapLogAdapter) WithGroup(name string) slog.Handler {
	return a
}
