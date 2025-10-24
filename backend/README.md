# GNX Backend

智能动漫生成系统后端服务

## 技术栈

- Go 1.21+
- Gin Web Framework
- GORM ORM
- PostgreSQL

## 目录结构

```
/backend/
├── main.go                 # 应用入口
├── config/                 # 配置文件
├── internal/              # 内部代码
│   ├── app/              # 应用初始化
│   ├── handlers/         # HTTP 处理器
│   ├── services/         # 业务逻辑层
│   ├── repositories/     # 数据访问层
│   ├── models/           # 数据模型
│   ├── middleware/       # 中间件
│   └── utils/            # 工具函数
├── pkg/                   # 可复用包
│   ├── ai/               # AI 服务集成
│   └── storage/          # 存储服务集成
└── migrations/           # 数据库迁移文件
```

## 开发指南

### 环境准备

1. 安装 Go 1.21+
2. 安装 PostgreSQL
3. 配置应用（支持两种方式，配置文件优先级更高）：
   - 方式一：复制 `config/config.yaml.example` 为 `config/config.yaml` 并配置（推荐）
   - 方式二：复制 `.env.example` 为 `.env` 并配置环境变量

### 运行

```bash
cd backend
go mod download
go run main.go
```

服务将在 `http://localhost:8080` 启动

## API 接口

详见 `/docs/api-docs.yaml`
