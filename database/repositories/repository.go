package repositories

import (
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Repository 基础存储库接口
返回: 基础存储库接口
*/
type Repository interface {
	// 基础接口方法，可根据需要扩展
}

/*
BaseMongoRepository MongoDB基础存储库
db: 数据库
返回: MongoDB基础存储库
*/
type BaseMongoRepository struct {
	DB         *mongo.Database
	Collection *mongo.Collection
}

/*
创建MongoDB基础存储库
db: 数据库
返回: MongoDB基础存储库
*/
func NewBaseMongoRepository(db *mongo.Database, collectionName string) *BaseMongoRepository {
	// 防御性编程：检查数据库连接是否为nil
	if db == nil {
		// 返回一个安全的基础仓库，其中Collection为nil
		return &BaseMongoRepository{
			DB:         nil,
			Collection: nil,
		}
	}

	return &BaseMongoRepository{
		DB:         db,
		Collection: db.Collection(collectionName),
	}
}

// RepositoryManager 存储库管理器
// 所有仓库的统一访问点
type RepositoryManager struct {
	mongoDB *mongo.Database
	User    UserRepository
	// 可以添加其他仓库...
}

// NewRepositoryManager 创建仓库管理器
func NewRepositoryManager(mongoDB *mongo.Database) *RepositoryManager {
	manager := &RepositoryManager{
		mongoDB: mongoDB,
	}

	// 初始化各个仓库
	if mongoDB != nil {
		// 使用MongoDB作为用户存储库的实现
		manager.User = NewUserRepository(mongoDB)
	} else {
		manager.User = &NullUserRepository{}
	}

	return manager
}
