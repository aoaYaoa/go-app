package repositories

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
MongoRepository MongoDB通用存储库
db: 数据库
collectionName: 集合名称
返回: MongoDB存储库
*/
type MongoRepository struct {
	db         *mongodb.Database
	collection *mongodb.Collection
}

/*
创建新的MongoDB存储库
db: 数据库
collectionName: 集合名称
返回: MongoDB存储库
*/
func NewMongoRepository(db *mongodb.Database, collectionName string) *MongoRepository {
	// 防御性编程：检查数据库连接是否为nil
	if db == nil {
		// 在实际应用中，应当避免传入空数据库连接
		// 可以在这里添加日志记录
		return &MongoRepository{
			db:         nil,
			collection: nil,
		}
	}

	return &MongoRepository{
		db:         db,
		collection: db.Collection(collectionName),
	}
}

/*
查找所有文档
filter: 查询条件
skip: 跳过数量
limit: 限制数量
sort: 排序
返回: 文档列表, 总数, 错误
*/
func (r *MongoRepository) FindAll(filter bson.M, skip, limit int64, sort bson.D) ([]bson.M, int64, error) {
	// 检查数据库连接和集合是否可用
	if r.db == nil || r.collection == nil {
		return nil, 0, fmt.Errorf("数据库连接不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 计算总数
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询选项
	opts := options.Find()
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if len(sort) > 0 {
		opts.SetSort(sort)
	}

	// 执行查询
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// 解析结果
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

/*
根据ID查找文档
id: 文档ID
返回: 文档, 错误
*/
func (r *MongoRepository) FindByID(id string) (bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的ID格式: %w", err)
	}

	var result bson.M
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, err
	}

	return result, nil
}

/*
根据条件查找单个文档
filter: 查询条件
返回: 文档, 错误
*/
func (r *MongoRepository) FindOne(filter bson.M) (bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result bson.M
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			return nil, fmt.Errorf("文档不存在")
		}
		return nil, err
	}

	return result, nil
}

/*
创建文档
document: 文档
返回: 文档ID, 错误
*/
func (r *MongoRepository) Create(document interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 确保创建和更新时间字段存在
	rv := reflect.ValueOf(document)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		now := time.Now()

		createdAtField := rv.FieldByName("CreatedAt")
		if createdAtField.IsValid() && createdAtField.CanSet() {
			createdAtField.Set(reflect.ValueOf(now))
		}

		updatedAtField := rv.FieldByName("UpdatedAt")
		if updatedAtField.IsValid() && updatedAtField.CanSet() {
			updatedAtField.Set(reflect.ValueOf(now))
		}
	}

	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return "", err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("无法获取插入的ID")
	}

	return id.Hex(), nil
}

/*
更新文档
id: 文档ID
update: 更新条件
返回: 错误
*/
func (r *MongoRepository) Update(id string, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的ID格式: %w", err)
	}

	// 添加更新时间
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	updateSet := update["$set"].(bson.M)
	updateSet["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("文档不存在")
	}

	return nil
}

/*
删除文档
id: 文档ID
返回: 错误
*/
func (r *MongoRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的ID格式: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("文档不存在")
	}

	return nil
}

/*
保存文档（创建或更新）
document: 文档
返回: 错误
*/
func (r *MongoRepository) Save(document interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rv := reflect.ValueOf(document)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// 获取ID字段
	idField := rv.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("文档没有ID字段")
	}

	id := idField.Interface().(primitive.ObjectID)
	filter := bson.M{"_id": id}

	// 设置更新时间
	now := time.Now()
	updatedAtField := rv.FieldByName("UpdatedAt")
	if updatedAtField.IsValid() && updatedAtField.CanSet() {
		updatedAtField.Set(reflect.ValueOf(now))
	}

	// 如果ID为空，则创建
	if id.IsZero() {
		// 设置创建时间
		createdAtField := rv.FieldByName("CreatedAt")
		if createdAtField.IsValid() && createdAtField.CanSet() {
			createdAtField.Set(reflect.ValueOf(now))
		}

		_, err := r.collection.InsertOne(ctx, document)
		if err != nil {
			return err
		}

		return nil
	}

	// 否则更新
	opts := options.FindOneAndReplace().SetUpsert(true)
	return r.collection.FindOneAndReplace(ctx, filter, document, opts).Err()
}
