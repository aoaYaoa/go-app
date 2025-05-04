package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 配置结构体，用于存储应用程序的所有配置信息
type Config struct {
	// Server 服务器相关配置
	Server struct {
		Port         string        `mapstructure:"SERVER_PORT"`          // 服务器监听端口
		Mode         string        `mapstructure:"SERVER_MODE"`          // 运行模式：debug/release
		ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`  // 读取超时时间
		WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"` // 写入超时时间
		IdleTimeout  time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`  // 空闲超时时间
	} `mapstructure:"server"`

	// Database 数据库相关配置
	Database struct {
		Host            string        `mapstructure:"DATABASE_HOST"`              // 数据库主机地址
		Port            int           `mapstructure:"DATABASE_PORT"`              // 数据库端口
		User            string        `mapstructure:"DATABASE_USER"`              // 数据库用户名
		Password        string        `mapstructure:"DATABASE_PASSWORD"`          // 数据库密码
		Name            string        `mapstructure:"DATABASE_NAME"`              // 数据库名称
		MaxIdleConns    int           `mapstructure:"DATABASE_MAX_IDLE_CONNS"`    // 最大空闲连接数
		MaxOpenConns    int           `mapstructure:"DATABASE_MAX_OPEN_CONNS"`    // 最大打开连接数
		ConnMaxLifetime time.Duration `mapstructure:"DATABASE_CONN_MAX_LIFETIME"` // 连接最大生命周期
	} `mapstructure:"database"`

	// MongoDB MongoDB数据库相关配置
	MongoDB struct {
		URI      string `mapstructure:"MONGODB_URI"`      // MongoDB连接URI
		Database string `mapstructure:"MONGODB_DATABASE"` // MongoDB数据库名称
		Username string `mapstructure:"MONGODB_USERNAME"` // MongoDB用户名
		Password string `mapstructure:"MONGODB_PASSWORD"` // MongoDB密码
	} `mapstructure:"mongodb"`

	// JWT JWT认证相关配置
	JWT struct {
		Secret string        `mapstructure:"JWT_SECRET"` // JWT密钥
		Expire time.Duration `mapstructure:"JWT_EXPIRE"` // JWT过期时间
	} `mapstructure:"jwt"`

	// Signature API签名相关配置
	Signature struct {
		AppKey    string        `mapstructure:"SIGNATURE_APP_KEY"`    // 应用id
		AppSecret string        `mapstructure:"SIGNATURE_APP_SECRET"` // 应用密钥
		Expire    time.Duration `mapstructure:"SIGNATURE_EXPIRE"`     // 签名过期时间
	} `mapstructure:"signature"`

	// CORS 跨域相关配置
	CORS struct {
		AllowOrigins     []string      `mapstructure:"CORS_ALLOW_ORIGINS"`     // 允许的源
		AllowCredentials bool          `mapstructure:"CORS_ALLOW_CREDENTIALS"` // 是否允许凭证
		MaxAge           time.Duration `mapstructure:"CORS_MAX_AGE"`           // 预检请求缓存时间
	} `mapstructure:"cors"`

	// Whitelist 白名单相关配置
	Whitelist struct {
		IPWhitelist         []string `mapstructure:"WHITELIST_IP"`          // IP白名单列表
		PathWhitelist       []string `mapstructure:"WHITELIST_PATH"`        // 路径白名单列表
		EnableIPWhitelist   bool     `mapstructure:"WHITELIST_IP_ENABLE"`   // 是否启用IP白名单
		EnablePathWhitelist bool     `mapstructure:"WHITELIST_PATH_ENABLE"` // 是否启用路径白名单
	} `mapstructure:"whitelist"`

	// Logger 日志相关配置
	Logger struct {
		Dir           string `mapstructure:"LOGGER_DIR"`            // 日志目录
		FileName      string `mapstructure:"LOGGER_FILENAME"`       // 日志文件名
		MaxSize       int    `mapstructure:"LOGGER_MAX_SIZE"`       // 单个日志文件最大大小(MB)
		MaxBackups    int    `mapstructure:"LOGGER_MAX_BACKUPS"`    // 最大保留旧日志文件数
		MaxAge        int    `mapstructure:"LOGGER_MAX_AGE"`        // 日志保留天数
		Compress      bool   `mapstructure:"LOGGER_COMPRESS"`       // 是否压缩旧日志文件
		ConsoleOutput bool   `mapstructure:"LOGGER_CONSOLE_OUTPUT"` // 是否输出到控制台
		RotateDaily   bool   `mapstructure:"LOGGER_ROTATE_DAILY"`   // 是否按天轮转日志
	} `mapstructure:"logger"`
}

// LoadConfig 加载配置文件
// 根据环境变量APP_ENV加载对应的配置文件（.env.test或.env.prod）
// 如果未设置APP_ENV，默认使用测试环境配置
func LoadConfig() *Config {
	// 获取环境变量
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "test" // 默认使用测试环境
	}

	// 设置配置文件
	viper.SetConfigName(".env." + env)
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic("无法读取配置文件: " + err.Error())
	}

	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic("无法解析配置文件: " + err.Error())
	}

	return &config
}
