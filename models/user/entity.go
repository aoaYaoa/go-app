package user

import (
	"time"
)

/*
* 实体模型指的是数据库中的表结构
* 用户实体模型
* 返回: 用户实体模型
 */
type User struct {
	ID        uint      `json:"id" bson:"id"`
	Username  string    `json:"username" bson:"username"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"password"`
	Nickname  string    `json:"nickname" bson:"nickname"`
	Avatar    string    `json:"avatar" bson:"avatar"`
	Status    int       `json:"status" bson:"status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	Deleted   bool      `json:"-" bson:"deleted"`
}

/*
返回用户表名
返回: 用户表名
*/
func (User) TableName() string {
	return "users" //mongodb 的集合名称 /mysql 的表名称
}
