package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-app/middleware"
	"go-app/models/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 集合名称常量
const (
	UserCollection = "users"
)

// InitMongoDB迁移 - 创建集合和索引
func MigrateDB() error {
	log.Println("开始MongoDB迁移...")

	if MongoDB == nil {
		return fmt.Errorf("MongoDB未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 初始化用户集合
	if err := setupUserCollection(ctx); err != nil {
		return fmt.Errorf("用户集合设置失败: %w", err)
	}

	// 添加默认管理员用户(如果不存在)
	if err := createDefaultAdmin(ctx); err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	log.Println("MongoDB迁移成功")
	return nil
}

// 设置用户集合和索引
func setupUserCollection(ctx context.Context) error {
	// 获取集合
	collection := MongoDB.Collection(UserCollection)

	// 创建索引
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}

	// 创建索引
	_, err := collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	return nil
}

// 创建默认管理员用户(如果不存在)
func createDefaultAdmin(ctx context.Context) error {
	collection := MongoDB.Collection(UserCollection)

	// 检查管理员是否已存在
	filter := bson.M{"username": "admin"}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("检查管理员用户失败: %w", err)
	}

	// 如果已存在管理员，则跳过
	if count > 0 {
		log.Println("管理员用户已存在，跳过创建")
		return nil
	}

	// 创建管理员密码哈希
	hashedPassword, err := middleware.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("管理员密码加密失败: %w", err)
	}

	// 创建管理员用户
	admin := user.User{
		ID:        1,
		Username:  "admin",
		Email:     "admin@example.com",
		Password:  hashedPassword,
		Nickname:  "管理员",
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 插入管理员用户
	_, err = collection.InsertOne(ctx, admin)
	if err != nil {
		return fmt.Errorf("插入管理员用户失败: %w", err)
	}

	log.Println("成功创建管理员用户")
	return nil
}
