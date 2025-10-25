# GNX Backend

智能动漫生成系统后端服务

## 技术栈

- Go 1.21+
- Gin Web Framework
- GORM ORM
- PostgreSQL
- 七牛云存储
- OpenAI API

## 项目结构

```
/backend/
├── go.mod                  # Go模块定义
├── go.sum                  # 依赖版本锁定
├── main.go                 # 应用入口
├── config/                 # 配置管理
│   ├── config.go          # 配置加载
│   └── database.go        # 数据库配置
├── internal/              # 内部包
│   ├── app/              # 应用层
│   │   ├── server.go     # 服务器初始化
│   │   └── routes.go     # 路由配置
│   ├── handlers/         # HTTP处理器
│   │   ├── comic.go      # 漫画相关接口
│   │   ├── section.go    # 章节相关接口
│   │   └── tts.go        # TTS相关接口
│   ├── services/         # 业务逻辑层
│   ├── repositories/     # 数据访问层
│   ├── models/          # 数据模型
│   │   ├── comic.go     # 漫画模型
│   │   ├── section.go   # 章节模型
│   │   ├── storyboard.go # 分镜模型
│   │   └── role.go      # 角色模型
│   ├── middleware/      # 中间件
│   │   ├── cors.go      # CORS中间件
│   │   ├── logging.go   # 日志中间件
│   │   └── recovery.go  # 恢复中间件
│   └── utils/           # 工具函数
│       └── response.go  # 响应格式化
├── pkg/                 # 公共包
│   ├── ai/             # AI服务
│   │   └── openai.go   # OpenAI客户端
│   └── storage/        # 存储服务
│       └── qiniu.go    # 七牛云客户端
├── migrations/         # 数据库迁移
├── scripts/           # 脚本文件
├── docker/            # Docker配置
└── tests/             # 测试文件
```

## 环境变量

```bash
# 服务器配置
SERVER_PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gnx

# OpenAI 配置
OPENAI_API_KEY=your_openai_api_key

# 七牛云配置
QINIU_ACCESS_KEY=your_qiniu_access_key
QINIU_SECRET_KEY=your_qiniu_secret_key
QINIU_BUCKET=your_bucket_name
QINIU_DOMAIN=your_domain
```

## 快速开始

1. 安装依赖
```bash
go mod download
```

2. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，填入实际配置
```

3. 运行服务
```bash
go run main.go
```

## API文档

详见 `/docs/api-docs.yaml`

## 开发计划

- [ ] 完善数据库迁移脚本
- [ ] 实现AI服务集成
- [ ] 完善业务逻辑层
- [ ] 添加单元测试
- [ ] 添加集成测试
