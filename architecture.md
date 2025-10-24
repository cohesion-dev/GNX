# 智能动漫生成系统架构设计

## 1. 系统整体架构

### 1.1 架构概览
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用       │    │   后端服务        │    │   外部服务       │
│   (Next.js)     │    │   (Go + Gin)    │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • 用户界面       │    │ • API服务        │    │ • OpenAI API    │
│ • 状态管理       │◄──►│ • 业务逻辑       │◄──►│ • 七牛云存储      │
│ • 路由管理       │    │ • 数据访问       │    │ • 七牛云TTS       │
│ • 文件上传       │    │ • 异步任务       │    │                  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   数据库         │
                       │  (PostgreSQL)   │
                       └─────────────────┘
```

### 1.2 技术栈架构

#### 前端架构
```
┌─────────────────────────────────────────────────────────┐
│                    前端应用层                             │
├─────────────────────────────────────────────────────────┤
│  Next.js 14+ (CSR模式) + TypeScript + Tailwind CSS       │
├─────────────────────────────────────────────────────────┤
│  页面层: /pages, /components                             │
│  状态层: MobX Store                                      │
│  服务层: /apis (HTTP客户端)                               │
│  工具层: /utils, /hooks                                  │
└─────────────────────────────────────────────────────────┘
```
注：前端页面需要适配移动端（iPhone）

#### 后端架构
```
┌─────────────────────────────────────────────────────────┐
│                    后端服务层                             │
├─────────────────────────────────────────────────────────┤
│  Go + Gin + GORM + PostgreSQL                           │
├─────────────────────────────────────────────────────────┤
│  API层: /handlers (HTTP接口)                             │
│  业务层: /services (业务逻辑)                             │
│  数据层: /models (数据模型)                               │
│  工具层: /utils, /middleware                             │
└─────────────────────────────────────────────────────────┘
```

## 2. 系统模块设计

### 2.1 前端模块

#### 2.1.1 页面模块
- **漫画列表页面** (`/comic`)
  - 漫画列表组件
  - 上拉分页组件

- **漫画详情页面** (`/comic/detail/[id]`)
  - 漫画概览组件
  - 角色列表组件
  - 章节列表组件

- **新增漫画页面** (`/comic/add`)
  - 文件上传组件
  - 漫画风格提示词表单输入组件
  - 提交处理按钮

- **新增漫画章节页面**（`/comic/detail/[id]/section/add`）
  - 新增漫画章节组件

- **漫画播放页面** (`/comic/detail/[id]/read`)
  - 播放器组件
  - 进度控制组件
  - 设置面板组件

### 2.2 后端模块

#### 2.2.1 API 模块
```go
// 路由设计
func SetupRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        // 漫画相关
        api.GET("/comic", handlers.GetComics)
        api.POST("/comic", handlers.CreateComic)
        api.GET("/comic/:id", handlers.GetComic)
        api.GET("/comic/:id/roles", handlers.GetComicRoles)
        api.GET("/comic/:id/sections", handlers.GetComicSections)
        
        // 章节相关
        api.POST("/comic/:id/section", handlers.CreateSection)
        api.GET("/comic/:id/section/:section_id/content", handlers.GetSectionContent)
        api.GET("/comic/:id/section/:section_id/storyboards", handlers.GetStoryboards)
        
        // TTS相关
        api.GET("/tts/:storyboard_tts_id", handlers.GetTTSAudio)
    }
}
```

#### 2.2.2 业务逻辑模块
```go
// 服务层设计
type ComicService struct {
    db *gorm.DB
    aiService *AIService
    storageService *StorageService
}

func (s *ComicService) CreateComic(req *CreateComicRequest) (*Comic, error) {
    // 1. 创建漫画记录
    // 2. 启动AI分析任务
    // 3. 返回结果
}

func (s *ComicService) ProcessComicGeneration(comicID uint) error {
    // 1. 分析小说内容
    // 2. 生成角色信息
    // 3. 生成分镜内容
    // 4. 生成图片和音频
}
```

#### 2.2.3 AI 服务模块
```go
// AI 服务设计
type AIService struct {
    openaiClient *openai.Client
    qiniuTTS     *qiniu.TTSService
}

func (s *AIService) AnalyzeNovel(content string) (*NovelAnalysis, error) {
    // 使用 OpenAI 分析小说内容
}

func (s *AIService) GenerateCharacter(character *Character) error {
    // 生成角色头像和简介
}

func (s *AIService) GenerateStoryboard(section *Section) error {
    // 生成分镜图片
}
```

## 3. 数据流设计

### 3.1 漫画创建流程
```
用户上传小说 → 前端验证 → 后端接收 → 创建漫画记录 → 启动AI生成漫画任务 → 返回漫画ID
```

### 3.2 AI生成流程
```
小说内容 → AI分析 → 角色信息 → 角色图片生成 → 分镜分析 → 分镜图片生成 → TTS角色信息生成 → 更新状态
```

## 4. 部署架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端开发      │    │   后端开发      │    │   数据库        │
│   (localhost)   │    │   (localhost)   │    │  (PostgreSQL)   │
│   Port: 3000    │    │   Port: 8080    │    │   Port: 5432    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```


