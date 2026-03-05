package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Env 运行环境类型
type Env string

const (
	Dev  Env = "dev"
	Prod Env = "prod"
)

// Config 日志配置结构体
type Config struct {
	Env           Env    `mapstructure:"env" yaml:"env"`                       // 运行环境：dev/prod
	Level         string `mapstructure:"level" yaml:"level"`                   // 日志级别：debug/info/warn/error/fatal/panic
	FileName      string `mapstructure:"file-name" yaml:"file-name"`           // 日志文件路径
	MaxSize       int    `mapstructure:"max-size" yaml:"max-size"`             // 单个日志文件最大大小（MB）
	MaxBackups    int    `mapstructure:"max-backups" yaml:"max-backups"`       // 保留备份文件数
	MaxAge        int    `mapstructure:"max-age" yaml:"max-age"`               // 日志文件保留天数
	Compress      bool   `mapstructure:"compress" yaml:"compress"`             // 是否压缩备份日志
	ShowCaller    bool   `mapstructure:"show-caller" yaml:"show-caller"`       // 是否显示调用者信息
	OutputConsole bool   `mapstructure:"output-console" yaml:"output-console"` // 是否同时输出到控制台
}

// DefaultConfig 返回默认日志配置
func DefaultConfig() Config {
	return Config{
		Env:           Dev,
		Level:         "info",
		FileName:      "./logs/app.log",
		MaxSize:       100,
		MaxBackups:    10,
		MaxAge:        30,
		Compress:      true,
		ShowCaller:    true,
		OutputConsole: true,
	}
}

// 全局日志实例
var (
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger

	// 额外的日志写入器
	extraWriters   []io.Writer
	extraWritersMu sync.RWMutex
	currentConfig  Config
)

// Init 初始化全局日志模块
func Init(cfg Config) error {
	currentConfig = cfg
	extraWriters = nil

	zLogger, err := New(cfg)
	if err != nil {
		return err
	}

	logger = zLogger
	sugarLogger = logger.Sugar()

	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger)

	return nil
}

// AddWriter 添加额外的日志写入器
func AddWriter(w io.Writer) {
	extraWritersMu.Lock()
	extraWriters = append(extraWriters, w)
	extraWritersMu.Unlock()

	rebuildLogger()
}

// RemoveWriter 移除额外的日志写入器
func RemoveWriter(w io.Writer) {
	extraWritersMu.Lock()
	for i, writer := range extraWriters {
		if writer == w {
			extraWriters = append(extraWriters[:i], extraWriters[i+1:]...)
			break
		}
	}
	extraWritersMu.Unlock()

	rebuildLogger()
}

// rebuildLogger 重建 logger 以包含新的写入器
func rebuildLogger() {
	zLogger, err := New(currentConfig)
	if err != nil {
		return
	}

	logger = zLogger
	sugarLogger = logger.Sugar()

	zap.ReplaceGlobals(logger)
}

// New 创建新的 Logger 实例
func New(cfg Config) (*zap.Logger, error) {
	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level failed: %w", err)
	}

	// 配置日志轮转
	hook := &lumberjack.Logger{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	// 构建日志写入器
	var writeSyncers []zapcore.WriteSyncer
	writeSyncers = append(writeSyncers, zapcore.AddSync(hook))
	if cfg.OutputConsole {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}

	// 添加额外的写入器
	extraWritersMu.RLock()
	for _, w := range extraWriters {
		writeSyncers = append(writeSyncers, zapcore.AddSync(w))
	}
	extraWritersMu.RUnlock()

	multiWriter := zapcore.NewMultiWriteSyncer(writeSyncers...)

	// 根据环境配置编码器
	encoder := buildEncoder(cfg.Env)

	// 构建 Zap 核心
	core := zapcore.NewCore(encoder, multiWriter, zap.NewAtomicLevelAt(level))

	// 构建 Logger 选项
	opts := []zap.Option{
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
		zap.AddStacktrace(zap.ErrorLevel),
	}
	if cfg.ShowCaller {
		opts = append(opts, zap.AddCaller())
	}

	zapLog := zap.New(core, opts...)
	return zapLog, nil
}

// buildEncoder 根据环境构建编码器
func buildEncoder(env Env) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
	}

	switch env {
	case Dev:
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	case Prod:
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		return zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
}

// L 返回全局 Logger 实例
func L() *zap.Logger {
	return logger
}

// S 返回全局 SugaredLogger 实例
func S() *zap.SugaredLogger {
	return sugarLogger
}

// Sync 刷新全局日志缓冲区
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// ======================== 快捷方法 ========================

// Debug 输出 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Info 输出 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// Warn 输出 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// Error 输出 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// Fatal 输出 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

// Panic 输出 Panic 级别日志并 panic
func Panic(msg string, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Panic(msg, fields...)
}

// Debugf 格式化输出 Debug 日志
func Debugf(template string, args ...any) {
	sugarLogger.WithOptions(zap.AddCallerSkip(1)).Debugf(template, args...)
}

// Infof 格式化输出 Info 日志
func Infof(template string, args ...any) {
	sugarLogger.WithOptions(zap.AddCallerSkip(1)).Infof(template, args...)
}

// Warnf 格式化输出 Warn 日志
func Warnf(template string, args ...any) {
	sugarLogger.WithOptions(zap.AddCallerSkip(1)).Warnf(template, args...)
}

// Errorf 格式化输出 Error 日志
func Errorf(template string, args ...any) {
	sugarLogger.WithOptions(zap.AddCallerSkip(1)).Errorf(template, args...)
}

// Fatalf 格式化输出 Fatal 日志并退出
func Fatalf(template string, args ...any) {
	sugarLogger.WithOptions(zap.AddCallerSkip(1)).Fatalf(template, args...)
}

// Panicf 格式化输出 Panic 日志并 panic
func Panicf(template string, args ...any) {
	sugarLogger.WithOptions(zap.AddCallerSkip(1)).Panicf(template, args...)
}

// With 添加结构化字段，返回新的 Logger
func With(fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}
