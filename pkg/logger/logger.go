package logger

import (
	"os"
	"strings"

	"github.com/lvtao/go-gin-api-admin/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func Init() error {
	// 日志级别（使用默认值）
	levelStr := "info"
	logFormat := "console"
	appSettings := config.GetAppSettings()
	if appSettings != nil {
		if l, ok := appSettings["logLevel"].(string); ok {
			levelStr = l
		}
		if f, ok := appSettings["logFormat"].(string); ok {
			logFormat = f
		}
	}

	var level zapcore.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if strings.ToLower(logFormat) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 输出到标准输出
	writeSyncer := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(encoder, writeSyncer, level)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Sugar = Logger.Sugar()

	return nil
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
