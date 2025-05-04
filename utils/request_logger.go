package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	requestLogger *RequestLogger
	reqLogOnce    sync.Once
)

// RequestLogger 专门用于记录HTTP请求的日志器
type RequestLogger struct {
	config LogConfig
	writer *lumberjack.Logger
	mutex  sync.Mutex
}

// RequestLog 请求日志结构
type RequestLog struct {
	Time      time.Time              `json:"time"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Query     string                 `json:"query"`
	Status    int                    `json:"status"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
	LatencyMs float64                `json:"latency_ms"`
	RequestID string                 `json:"request_id,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Params    map[string]string      `json:"params,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
	ExtraInfo map[string]interface{} `json:"extra_info,omitempty"`
}

// InitRequestLogger 初始化请求日志记录器
func InitRequestLogger(config LogConfig) {
	reqLogOnce.Do(func() {
		// 确保日志目录存在
		logDir := config.LogDir
		if logDir == "" {
			logDir = "logs/requests" // 默认请求日志目录
		} else {
			logDir = filepath.Join(logDir, "requests")
		}

		if err := os.MkdirAll(logDir, 0755); err != nil {
			Error("无法创建请求日志目录", zap.Error(err))
			return
		}

		// 初始化请求日志记录器
		requestLogger = &RequestLogger{
			config: config,
			mutex:  sync.Mutex{},
		}

		// 启动一个goroutine，每天更新日志文件名
		if config.RotateDaily {
			go func() {
				for {
					// 更新日志文件名
					requestLogger.updateWriter()

					// 计算下一天的时间
					now := time.Now()
					tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
					duration := tomorrow.Sub(now)

					// 等待到明天0点
					time.Sleep(duration)
				}
			}()
		} else {
			requestLogger.updateWriter()
		}

		Info("请求日志系统初始化成功",
			zap.String("日志目录", logDir),
			zap.Bool("按天轮转", config.RotateDaily),
		)
	})
}

// 更新日志写入器，根据当前日期生成日志文件名
func (rl *RequestLogger) updateWriter() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// 获取当前日期
	var logFilename string

	// 根据配置决定是否使用日期命名
	if rl.config.RotateDaily {
		today := time.Now().Format("2006-01-02")
		logFilename = filepath.Join(rl.config.LogDir, "requests", fmt.Sprintf("requests-%s.log", today))
	} else {
		logFilename = filepath.Join(rl.config.LogDir, "requests", "requests.log")
	}

	// 创建或更新写入器
	rl.writer = &lumberjack.Logger{
		Filename:   logFilename,
		MaxSize:    rl.config.MaxSize,
		MaxBackups: rl.config.MaxBackups,
		MaxAge:     rl.config.MaxAge,
		Compress:   rl.config.Compress,
	}
}

// LogRequest 记录请求日志
func LogRequest(reqLog RequestLog) {
	if requestLogger == nil {
		// 如果请求日志器未初始化，使用默认配置初始化
		InitRequestLogger(defaultLogConfig)
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(reqLog)
	if err != nil {
		Error("请求日志序列化失败", zap.Error(err))
		return
	}

	// 添加换行符
	jsonData = append(jsonData, '\n')

	// 写入日志
	requestLogger.mutex.Lock()
	defer requestLogger.mutex.Unlock()

	// 确保writer已初始化
	if requestLogger.writer == nil {
		requestLogger.updateWriter()
	}

	// 写入日志数据
	if _, err := requestLogger.writer.Write(jsonData); err != nil {
		Error("请求日志写入失败", zap.Error(err))
	}
}
