package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
	once        sync.Once
)

// LogConfig 日志配置
type LogConfig struct {
	LogDir        string // 日志目录
	LogFileName   string // 日志文件名
	MaxSize       int    // 单个日志文件最大大小，单位MB
	MaxBackups    int    // 最大保留旧日志文件数
	MaxAge        int    // 日志文件保留天数
	Compress      bool   // 是否压缩旧日志文件
	ConsoleOutput bool   // 是否输出到控制台
	RotateDaily   bool   // 是否按天轮转
}

// 默认日志配置
var defaultLogConfig = LogConfig{
	LogDir:        "logs",
	LogFileName:   "app.log",
	MaxSize:       100,
	MaxBackups:    10,
	MaxAge:        30,
	Compress:      true,
	ConsoleOutput: true,
	RotateDaily:   true, // 默认开启按天轮转
}

// InitLogger 初始化日志，使用默认配置
func InitLogger() {
	InitLoggerWithConfig(defaultLogConfig)
}

// InitLoggerWithConfig 使用自定义配置初始化日志
func InitLoggerWithConfig(config LogConfig) {
	once.Do(func() {
		// 确保日志目录存在
		if err := os.MkdirAll(config.LogDir, 0755); err != nil {
			panic("无法创建日志目录: " + err.Error())
		}

		// 配置编码器
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
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

		// 创建JSON编码器
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

		// 日志级别
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		})

		// 获取当前日期
		var logFilename, errorLogFilename string

		if config.RotateDaily {
			// 加入日期到文件名中，实现按日期归档
			today := time.Now().Format("2006-01-02")
			logFilename = filepath.Join(config.LogDir, fmt.Sprintf("%s.log", today))
			errorLogFilename = filepath.Join(config.LogDir, fmt.Sprintf("%s_error.log", today))
		} else {
			logFilename = filepath.Join(config.LogDir, "info_"+config.LogFileName)
			errorLogFilename = filepath.Join(config.LogDir, "error_"+config.LogFileName)
		}

		// 常规日志文件
		infoLogFile := &lumberjack.Logger{
			Filename:   logFilename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		// 错误日志文件
		errorLogFile := &lumberjack.Logger{
			Filename:   errorLogFilename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		// 将文件WriteSyncer包装成zapcore.WriteSyncer
		infoFileWriter := zapcore.AddSync(infoLogFile)
		errorFileWriter := zapcore.AddSync(errorLogFile)

		// 构建日志核心
		var cores []zapcore.Core

		// 文件日志输出
		cores = append(cores,
			zapcore.NewCore(jsonEncoder, errorFileWriter, highPriority),
			zapcore.NewCore(jsonEncoder, infoFileWriter, lowPriority),
		)

		// 控制台日志输出(可选)
		if config.ConsoleOutput {
			consoleDebugging := zapcore.Lock(os.Stdout)
			consoleErrors := zapcore.Lock(os.Stderr)
			cores = append(cores,
				zapcore.NewCore(jsonEncoder, consoleErrors, highPriority),
				zapcore.NewCore(jsonEncoder, consoleDebugging, lowPriority),
			)
		}

		// 合并所有日志输出
		core := zapcore.NewTee(cores...)

		// 创建日志记录器，添加调用信息
		logger = zap.New(core,
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)

		// 创建糖化记录器
		sugarLogger = logger.Sugar()

		// 记录日志初始化成功
		logger.Info("日志系统初始化成功",
			zap.String("日志目录", config.LogDir),
			zap.String("日志文件名", config.LogFileName),
			zap.Bool("按天轮转", config.RotateDaily),
		)
	})
}

// GetLogger 获取日志记录器
func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

// GetSugarLogger 获取糖化日志记录器
func GetSugarLogger() *zap.SugaredLogger {
	if sugarLogger == nil {
		InitLogger()
	}
	return sugarLogger
}

// Debug logs a message at DebugLevel
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs a message at InfoLevel
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a message at WarnLevel
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a message at FatalLevel
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sync 同步日志缓冲区到文件
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
