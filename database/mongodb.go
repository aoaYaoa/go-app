package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go-app/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB 全局MongoDB客户端
var MongoDB *mongo.Database

// MongoClient 全局Mongo客户端
var MongoClient *mongo.Client

// MongoManager MongoDB管理器
type MongoManager struct {
	Client *mongo.Client
	DB     *mongo.Database
	Config *config.Config
}

// NewMongoManager 创建新的MongoDB管理器
func NewMongoManager(cfg *config.Config) *MongoManager {
	return &MongoManager{
		Config: cfg,
	}
}

// InitMongoDB 初始化MongoDB连接
func InitMongoDB(cfg *config.Config) (*mongo.Database, error) {
	log.Println("正在连接MongoDB...")

	// 处理空配置
	if cfg == nil {
		cfg = &config.Config{}
	}

	// 直接从环境变量读取 MongoDB URI
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		// 如果环境变量不存在，则使用配置
		uri = cfg.MongoDB.URI
		// 如果配置也不存在，使用默认值
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}
	}

	// 直接从环境变量读取数据库名
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		// 如果环境变量不存在，则使用配置
		dbName = cfg.MongoDB.Database
		// 如果配置也不存在，使用默认值
		if dbName == "" {
			dbName = "go_app"
		}
	}

	log.Printf("正在连接到 MongoDB: %s, 数据库: %s", uri, dbName)

	// 创建连接上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置客户端选项 - 不使用身份验证
	clientOptions := options.Client().ApplyURI(uri)

	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("无法连接MongoDB: %w", err)
	}

	// 检查连接
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("MongoDB连接测试失败: %w", err)
	}

	// 设置全局客户端
	MongoClient = client

	// 设置数据库
	db := client.Database(dbName)

	// 设置全局数据库
	MongoDB = db

	log.Println("MongoDB连接成功")
	return db, nil
}

// CloseMongoDB 关闭MongoDB连接
func CloseMongoDB() error {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := MongoClient.Disconnect(ctx); err != nil {
			return fmt.Errorf("关闭MongoDB连接失败: %w", err)
		}
		log.Println("MongoDB连接已关闭")
	}
	return nil
}

// GetCollection 获取MongoDB集合
func GetCollection(name string) *mongo.Collection {
	if MongoDB == nil {
		log.Println("警告: 尝试在MongoDB未初始化时获取集合")
		return nil
	}
	return MongoDB.Collection(name)
}

// InitMongoManager 初始化MongoDB管理器
func (m *MongoManager) InitMongoManager() error {
	db, err := InitMongoDB(m.Config)
	if err != nil {
		return err
	}

	m.DB = db
	m.Client = MongoClient

	return nil
}

// Close 关闭MongoDB连接
func (m *MongoManager) Close() error {
	return CloseMongoDB()
}

// Collection 获取集合
func (m *MongoManager) Collection(name string) *mongo.Collection {
	return m.DB.Collection(name)
}
