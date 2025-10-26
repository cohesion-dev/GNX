# GNX Backend (New)

全新设计的后端服务，完全按照 `docs/api-docs.md` API 文档实现。

## 架构设计

### 技术栈
- **语言**: Go 1.22+
- **框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **存储**: 七牛云对象存储
- **AI服务**: 基于 `ai/gnxaigc` 包

### 目录结构
```
backend_new/
├── cmd/                    # 应用入口
│   └── main.go
├── config/                 # 配置管理
│   └── config.go
├── internal/               # 内部代码
│   ├── models/            # 数据模型
│   │   ├── comic.go
│   │   ├── role.go
│   │   ├── section.go
│   │   ├── page.go
│   │   └── page_detail.go
│   ├── repositories/      # 数据访问层
│   │   ├── comic_repository.go
│   │   ├── role_repository.go
│   │   ├── section_repository.go
│   │   └── page_repository.go
│   ├── services/          # 业务逻辑层
│   │   ├── comic_service.go
│   │   ├── section_service.go
│   │   ├── image_service.go
│   │   └── tts_service.go
│   ├── handlers/          # HTTP 处理器
│   │   ├── comic_handler.go
│   │   ├── section_handler.go
│   │   ├── image_handler.go
│   │   └── tts_handler.go
│   ├── middleware/        # 中间件
│   │   ├── cors.go
│   │   ├── logging.go
│   │   └── recovery.go
│   ├── utils/             # 工具函数
│   │   └── response.go
│   └── app/               # 应用组装
│       ├── app.go
│       └── router.go
└── pkg/                    # 公共包
    ├── database/          # 数据库连接
    ├── storage/           # 七牛云存储
    └── aigc/              # AI 服务封装
```

## API 端点

所有 API 端点严格按照 `docs/api-docs.md` 实现：

### 漫画管理
- `GET /comics/` - 获取漫画列表
- `POST /comics/` - 创建新漫画（上传小说文件）
- `GET /comics/{comic_id}/` - 获取漫画详情

### 章节管理
- `POST /comics/{comic_id}/sections/` - 创建新章节
- `GET /comics/{comic_id}/sections/{section_id}/` - 获取章节详情

### 资源访问
- `GET /images/{image_id}/url` - 获取图片临时URL
- `GET /tts/{tts_id}` - 获取TTS音频流

## 数据模型

### Comic (漫画)
- ID、标题、用户提示词
- 封面图片ID、背景图片ID
- 状态（pending/completed/failed）

### ComicRole (角色)
- 名称、简介、性别、年龄
- 角色图片ID
- 语音配置（VoiceName, VoiceType）

### ComicSection (章节)
- 标题、索引、内容
- 状态（pending/completed/failed）

### ComicPage (页面)
- 对应API文档中的 `pages`
- 包含图片生成提示词

### ComicPageDetail (页面详情)
- 对应API文档中的 `details`
- 文字内容、关联角色
- ID 同时作为 TTS 标识符

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
DB_SSLMODE=disable

# 七牛云存储配置
QINIU_ACCESS_KEY=your_access_key
QINIU_SECRET_KEY=your_secret_key
QINIU_BUCKET=your_bucket
QINIU_DOMAIN=your_domain

# OpenAI 配置
OPENAI_API_KEY=your_api_key
OPENAI_BASE_URL=https://openai.qiniu.com/v1
OPENAI_IMAGE_MODEL=gemini-2.5-flash-image
OPENAI_LANGUAGE_MODEL=deepseek/deepseek-v3.1-terminus
```

## 运行方式

```bash
# 安装依赖
cd backend_new
go mod tidy

# 运行服务
go run cmd/main.go
```

## AI 集成流程

### 创建漫画流程
1. 接收小说文件和基本信息
2. 调用 `SummaryChapter` 分析小说，提取角色和分镜
3. 为每个角色生成概念图（`GenerateImageByText`）
4. 生成封面和背景图
5. 更新漫画状态为 completed

### 创建章节流程
1. 接收章节标题和内容
2. 加载已有角色信息
3. 调用 `SummaryChapter` 生成章节分镜
4. 为每页生成图片（`GenerateImageByText`）
5. 创建页面和详情记录
6. 更新章节状态为 completed

### TTS 生成流程
1. 接收 detail_id（即 tts_id）
2. 查找对应的文字内容和角色信息
3. 调用 `TextToSpeechSimple` 实时生成音频
4. 直接返回音频流

## 与旧后端的区别

1. **API路径**: 使用 `/comics/` 而不是 `/api/comic`
2. **响应格式**: 统一的 `{code, message, data}` 格式
3. **图片管理**: 返回图片ID，通过 `/images/{id}/url` 获取临时URL
4. **数据结构**: 简化为 Page + Detail，而不是 Storyboard + Panel + Segment
5. **TTS处理**: 实时生成而不是预生成存储

## 注意事项

1. AI 处理是异步的，创建漫画/章节后状态为 pending
2. 图片URL有效期为1小时，前端需要定期刷新
3. TTS音频是实时生成的，可能有延迟
4. 所有ID在API中以字符串形式返回
