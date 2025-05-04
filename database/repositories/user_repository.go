/**
* @Author: gcl
* @Date: 2025-04-29 15:25:00
 * @Last Modified by: mikey.zhaopeng
 * @Last Modified time: 2025-05-04 12:25:56
* @Description: 用户存储库实现,基于MongoDB
*/
package repositories

import (
	"context"
	"fmt"
	"time"

	"go-app/models/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 集合名称常量
const UserCollection = "users"

// UserRepository 用户存储库接口
type UserRepository interface {
	FindAll(page, pageSize int, conditions map[string]interface{}) ([]user.User, int64, error)
	FindByID(id uint) (*user.User, error)
	FindByUsername(username string) (*user.User, error)
	FindByEmail(email string) (*user.User, error)
	Create(user *user.User) error
	Update(user *user.User) error
	Delete(id uint) error
}

// MongoUserRepository MongoDB用户存储库实现
type MongoUserRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewUserRepository 创建新的用户存储库
func NewUserRepository(db *mongo.Database) UserRepository {
	if db == nil {
		return &NullUserRepository{}
	}

	return &MongoUserRepository{
		db:         db,
		collection: db.Collection(UserCollection),
	}
}

// FindAll 查找所有用户
func (r *MongoUserRepository) FindAll(page, pageSize int, conditions map[string]interface{}) ([]user.User, int64, error) {
	// 处理分页
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// 构建查询条件
	filter := bson.M{}

	// 添加状态过滤
	if status, ok := conditions["status"]; ok && status != nil {
		filter["status"] = status
	}

	// 添加关键词搜索
	if keyword, ok := conditions["keyword"].(string); ok && keyword != "" {
		// 使用$or操作符实现多字段搜索
		filter["$or"] = []bson.M{
			{"username": bson.M{"$regex": keyword, "$options": "i"}},
			{"email": bson.M{"$regex": keyword, "$options": "i"}},
			{"nickname": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

	// 设置排序方式：按创建时间降序
	sort := bson.D{{"created_at", -1}}

	// 获取上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 计算总记录数
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("计算用户总数失败: %w", err)
	}

	// 查询选项
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sort)

	// 执行查询
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var users []user.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, fmt.Errorf("解析用户列表失败: %w", err)
	}

	return users, count, nil
}

// FindByID 根据ID查找用户
func (r *MongoUserRepository) FindByID(id uint) (*user.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &u, nil
}

// FindByUsername 根据用户名查找用户
func (r *MongoUserRepository) FindByUsername(username string) (*user.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &u, nil
}

// FindByEmail 根据邮箱查找用户
func (r *MongoUserRepository) FindByEmail(email string) (*user.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &u, nil
}

// Create 创建用户
func (r *MongoUserRepository) Create(u *user.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置创建和更新时间
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	// 如果ID未设置，生成一个
	if u.ID == 0 {
		u.ID = generateUserID()
	}

	_, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// Update 更新用户
func (r *MongoUserRepository) Update(u *user.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 更新更新时间
	u.UpdatedAt = time.Now()

	filter := bson.M{"id": u.ID}
	update := bson.M{"$set": u}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// Delete 删除用户
func (r *MongoUserRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// 生成用户ID - 简单实现
func generateUserID() uint {
	// 基于当前时间戳生成ID
	// 实际应用中可以使用更复杂的ID生成算法
	return uint(time.Now().UnixNano() / 1000000)
}

// NullUserRepository 空用户存储库实现（空对象模式）
// 当数据库不可用时提供一个不会崩溃的实现
type NullUserRepository struct{}

// FindAll 查找所有用户 - 空实现
func (r *NullUserRepository) FindAll(page, pageSize int, conditions map[string]interface{}) ([]user.User, int64, error) {
	return []user.User{}, 0, fmt.Errorf("MongoDB数据库不可用，无法查询用户")
}

// FindByID 根据ID查找用户 - 空实现
func (r *NullUserRepository) FindByID(id uint) (*user.User, error) {
	return nil, fmt.Errorf("MongoDB数据库不可用，无法查询用户")
}

// FindByUsername 根据用户名查找用户 - 空实现
func (r *NullUserRepository) FindByUsername(username string) (*user.User, error) {
	return nil, fmt.Errorf("MongoDB数据库不可用，无法查询用户")
}

// FindByEmail 根据邮箱查找用户 - 空实现
func (r *NullUserRepository) FindByEmail(email string) (*user.User, error) {
	return nil, fmt.Errorf("MongoDB数据库不可用，无法查询用户")
}

// Create 创建用户 - 空实现
func (r *NullUserRepository) Create(u *user.User) error {
	return fmt.Errorf("MongoDB数据库不可用，无法创建用户")
}

// Update 更新用户 - 空实现
func (r *NullUserRepository) Update(u *user.User) error {
	return fmt.Errorf("MongoDB数据库不可用，无法更新用户")
}

// Delete 删除用户 - 空实现
func (r *NullUserRepository) Delete(id uint) error {
	return fmt.Errorf("MongoDB数据库不可用，无法删除用户")
}
