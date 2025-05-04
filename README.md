# Go Web应用框架

这是一个使用Go语言开发的Web应用框架，以MongoDB为数据库，集成了常用的功能和最佳实践。

## 项目结构

```
.
├── config/                 # 配置文件
│   └── config.go           # 配置结构和加载逻辑
├── controller/             # 控制器
│   ├── base.go             # 控制器基类
│   └── user/              
│       └── controller.go   # 用户控制器
├── database/               # 数据库相关
│   ├── migrate.go          # 数据库迁移
│   ├── mongodb.go          # MongoDB初始化
│   └── repositories/       # 数据访问层
│       ├── repository.go   # 存储库基类
│       └── user_repository.go # 用户存储库
├── middleware/             # 中间件
│   ├── cors.go             # CORS中间件
│   ├── error.go            # 错误处理中间件
│   ├── jwt.go              # JWT认证中间件
│   ├── logger.go           # 日志中间件
│   ├── middleware.go       # 中间件管理
│   ├── password.go         # 密码处理中间件
│   ├── signature.go        # API签名验证中间件
│   ├── validator.go        # 参数验证中间件
│   └── whitelist.go        # 白名单中间件
├── models/                 # 模型定义
│   ├── common/             # 通用模型
│   └── user/               # 用户相关模型
│       ├── entity.go       # 用户实体
│       ├── request.go      # 用户请求模型
│       └── response.go     # 用户响应模型
├── router/                 # 路由配置
│   ├── index.go            # 路由主文件
│   └── user.go             # 用户路由
├── service/                # 业务逻辑层
│   └── user_service.go     # 用户服务
├── utils/                  # 工具函数
│   ├── logger.go           # 日志工具
│   └── request_logger.go   # 请求日志工具
├── .env                    # 环境变量配置
├── go.mod                  # Go模块定义
├── go.sum                  # Go依赖校验
├── main.go                 # 主程序入口
└── README.md               # 项目说明
```

## 功能特性

- RESTful API设计
- JWT认证和授权
- MongoDB数据库支持
- CORS跨域支持
- API签名验证机制
- 结构化日志系统
  - 按天分割的应用日志
  - 按天分割的独立请求日志
- 丰富的中间件
  - 错误处理
  - IP和路径白名单
  - 请求签名验证
- 高性能的数据访问层
- 多环境配置支持
- 安全的密码存储与验证

## 技术栈

- [Gin](https://github.com/gin-gonic/gin) - 高性能Web框架
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - MongoDB官方驱动
- [Zap](https://github.com/uber-go/zap) - 结构化、高性能日志库
- [Lumberjack](https://github.com/natefinch/lumberjack) - 日志轮转和管理
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Token认证
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Validator](https://github.com/go-playground/validator) - 请求参数验证

## 开始使用

### 前提条件

- Go 1.18 或更高版本
- MongoDB 4.0 或更高版本

### 安装

1. 克隆代码库:

```bash
git clone https://github.com/username/go-app.git
cd go-app
```

2. 安装依赖:

```bash
go mod download
```

3. 配置环境:

创建或编辑 `.env` 文件，设置以下主要配置项：

```env
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# MongoDB配置
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=go_app

# JWT配置
JWT_SECRET=your_jwt_secret
JWT_EXPIRE=24h

# 日志配置
LOGGER_DIR=logs
LOGGER_ROTATE_DAILY=true

# API签名配置
SIGNATURE_APP_KEY=your_app_key
SIGNATURE_APP_SECRET=your_app_secret
SIGNATURE_EXPIRE=300s
```

4. 启动应用:

```bash
go run main.go
```

应用将在`http://localhost:8080`启动，可通过`/ping`接口测试服务是否正常运行。

## API接口

### 公开接口

- `POST /api/v1/users/register` - 用户注册
- `POST /api/v1/users/login` - 用户登录
- `GET /ping` - 健康检查

### 需要认证的接口

- `GET /api/v1/users` - 获取用户列表
- `GET /api/v1/users/:id` - 获取用户详情
- `DELETE /api/v1/users/:id` - 删除用户
- `GET /api/v1/users/profile` - 获取当前用户信息
- `PUT /api/v1/users/profile` - 更新当前用户信息
- `POST /api/v1/users/change-password` - 修改密码

## API签名验证

为确保API调用的安全性，本框架实现了请求签名验证机制。客户端需要按以下步骤生成签名：

1. 收集所有请求参数（GET参数或POST表单，不包括URL中的path参数）
2. 添加`app_key`、`timestamp`和`nonce`参数
3. 按参数名称字母顺序排序
4. 拼接为`key1=value1&key2=value2...&app_secret=YOUR_APP_SECRET`形式
5. 计算MD5哈希值作为签名
6. 将签名作为`sign`参数加入请求

示例代码（JavaScript）:
```javascript
function generateSignature(params, appSecret) {
  // 添加timestamp和nonce
  params.timestamp = Math.floor(Date.now() / 1000);
  params.nonce = Math.random().toString(36).substring(2);
  
  // 按键名排序
  const keys = Object.keys(params).sort();
  
  // 构建签名字符串
  let signStr = '';
  keys.forEach(key => {
    signStr += `${key}=${params[key]}&`;
  });
  signStr += `app_secret=${appSecret}`;
  
  // 计算MD5
  return CryptoJS.MD5(signStr).toString();
}
```

## 日志系统

本框架实现了强大的日志系统，主要特点：

1. **应用日志**：记录应用运行状态、错误等信息
   - 按照日期自动归档（YYYY-MM-DD.log）
   - 按级别区分（info/error）
   - 结构化JSON格式，便于分析

2. **请求日志**：专门记录HTTP请求的详细信息
   - 按天自动生成日志文件（requests-YYYY-MM-DD.log）
   - 包含请求方法、路径、参数、客户端信息等
   - 记录请求处理时间（延迟）
   - JSON格式，方便ELK等系统处理

日志目录结构：
```
logs/
├── 2025-05-04.log             # 应用日志
├── 2025-05-04_error.log       # 错误日志
└── requests/
    └── requests-2025-05-04.log # 请求日志
```

## 部署说明

1. 构建生产环境二进制文件
```bash
go build -o app
```

2. 创建配置文件
```bash
cp .env.example .env.prod
```
编辑`.env.prod`文件，设置生产环境配置

3. 运行应用
```bash
./app
```

## 安全建议

1. 配置文件中的敏感信息（如数据库密码、JWT密钥）应妥善保管
2. 生产环境应启用HTTPS
3. 定期更新密码和密钥
4. 使用强密码和足够长的密钥
5. 根据需要调整JWT过期时间

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

[MIT](LICENSE) 