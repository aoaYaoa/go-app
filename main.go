package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-app/config"
	"go-app/database"
	"go-app/database/repositories"
	"go-app/middleware"
	"go-app/router"
	"go-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("警告: .env文件未找到，使用系统环境变量")
	}

	// 加载配置
	cfg := config.LoadConfig()

	// 日志配置
	logFileName := "app.log"
	if cfg.Logger.FileName != "" {
		logFileName = cfg.Logger.FileName
	}

	maxSize := 100
	if cfg.Logger.MaxSize > 0 {
		maxSize = cfg.Logger.MaxSize
	}

	maxBackups := 10
	if cfg.Logger.MaxBackups > 0 {
		maxBackups = cfg.Logger.MaxBackups
	}

	maxAge := 30
	if cfg.Logger.MaxAge > 0 {
		maxAge = cfg.Logger.MaxAge
	}

	// 初始化日志
	utils.InitLoggerWithConfig(utils.LogConfig{
		LogDir:        "logs", // 默认日志目录
		LogFileName:   logFileName,
		MaxSize:       maxSize,
		MaxBackups:    maxBackups,
		MaxAge:        maxAge,
		Compress:      cfg.Logger.Compress,
		ConsoleOutput: true,
		RotateDaily:   true, // 强制按天轮转
	})

	// 初始化请求日志记录器
	utils.InitRequestLogger(utils.LogConfig{
		LogDir:      "logs", // 日志将保存在 logs/requests 目录下
		MaxSize:     maxSize,
		MaxBackups:  maxBackups,
		MaxAge:      maxAge,
		Compress:    cfg.Logger.Compress,
		RotateDaily: true, // 按天生成日志文件
	})

	// 确保日志在程序退出时正确刷新
	defer utils.Sync()

	utils.Info("应用程序启动")

	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化MongoDB连接
	mongoDb, err := database.InitMongoDB(cfg)
	if err != nil {
		utils.Error("MongoDB初始化失败", zap.Error(err))
		utils.Fatal("无法启动应用程序，MongoDB连接失败")
		return
	}

	// 执行MongoDB迁移
	// 暂时不执行迁移
	// if err := database.MigrateDB(); err != nil {
	// 	utils.Error("MongoDB迁移失败", zap.Error(err))
	// 	utils.Warn("将继续运行，但可能缺少一些必要的初始数据")
	// }

	// 创建存储库管理器，使用MongoDB
	repoManager := repositories.NewRepositoryManager(mongoDb)
	utils.Info("MongoDB初始化成功")

	// 创建Gin引擎
	r := gin.New()

	// 添加Recovery中间件
	r.Use(gin.Recovery())

	// 添加日志和错误处理中间件
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())

	// 添加CORS中间件
	r.Use(middleware.Cors(cfg))

	// 设置路由
	router.Setup(r, cfg, repoManager)

	// 配置服务器
	port := cfg.Server.Port
	if port == "" {
		port = "8080" // 使用默认端口
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		utils.Info(fmt.Sprintf("服务器启动于 http://localhost:%s", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Error("服务器运行出错", zap.Error(err))
		}
	}()

	// 等待信号
	<-quit
	utils.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.Error("服务器关闭出错", zap.Error(err))
	}

	utils.Info("服务器已关闭")
}
