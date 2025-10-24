# 数据库设计文档

## 1. 数据库概览

### 1.1 数据库选择
- **数据库类型**: PostgreSQL 15+
- **ORM框架**: GORM
- **连接池**: 支持连接池管理
- **索引策略**: 基于查询模式优化索引

### 1.2 数据库设计原则
- 遵循第三范式设计
- 合理使用索引提升查询性能
- 支持软删除机制
- 统一的时间戳字段

## 2. 数据表设计

### 2.1 动漫表 (comic)
```sql
CREATE TABLE comic (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL COMMENT '动漫名称',
    brief TEXT COMMENT '动漫简介',
    icon VARCHAR(500) COMMENT '动漫图标URL',
    bg VARCHAR(500) COMMENT '动漫背景图URL',
    user_prompt TEXT COMMENT '用户提示词',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '生成状态: pending, completed, failed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间'
);

-- 索引
CREATE INDEX idx_comic_status ON comic(status);
CREATE INDEX idx_comic_created_at ON comic(created_at);
```

### 2.2 动漫角色表 (comic_role)
```sql
CREATE TABLE comic_role (
    id SERIAL PRIMARY KEY,
    comic_id INTEGER NOT NULL REFERENCES comic(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL COMMENT '角色名称',
    image_url VARCHAR(500) COMMENT '角色图片URL',
    brief TEXT COMMENT '角色简介',
    voice VARCHAR(100) COMMENT '音色标识',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_comic_role_comic_id ON comic_role(comic_id);
CREATE INDEX idx_comic_role_name ON comic_role(name);
```

### 2.3 动漫章节表 (comic_section)
```sql
CREATE TABLE comic_section (
    id SERIAL PRIMARY KEY,
    comic_id INTEGER NOT NULL REFERENCES comic(id) ON DELETE CASCADE,
    index INTEGER NOT NULL COMMENT '章节序号',
    detail TEXT NOT NULL COMMENT '章节原文',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '生成状态: pending, completed, failed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_comic_section_comic_id ON comic_section(comic_id);
CREATE INDEX idx_comic_section_index ON comic_section(comic_id, index);
CREATE INDEX idx_comic_section_status ON comic_section(status);
```

### 2.4 动漫运镜表 (comic_storyboard)
```sql
CREATE TABLE comic_storyboard (
    id SERIAL PRIMARY KEY,
    section_id INTEGER NOT NULL REFERENCES comic_section(id) ON DELETE CASCADE,
    image_prompt TEXT COMMENT '运镜提示词',
    image_url VARCHAR(500) COMMENT '运镜图片URL',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_comic_storyboard_section_id ON comic_storyboard(section_id);
```

### 2.5 动漫运镜详情表 (comic_storyboard_detail)
```sql
CREATE TABLE comic_storyboard_detail (
    id SERIAL PRIMARY KEY,
    storyboard_id INTEGER NOT NULL REFERENCES comic_storyboard(id) ON DELETE CASCADE,
    detail TEXT COMMENT '原文内容',
    role_id INTEGER REFERENCES comic_role(id) ON DELETE SET NULL,
    tts_url VARCHAR(500) COMMENT 'TTS音频URL',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_comic_storyboard_detail_storyboard_id ON comic_storyboard_detail(storyboard_id);
CREATE INDEX idx_comic_storyboard_detail_role_id ON comic_storyboard_detail(role_id);
```

### 2.6 动漫角色运镜关系表 (comic_role_storyboard)
```sql
CREATE TABLE comic_role_storyboard (
    id SERIAL PRIMARY KEY,
    storyboard_id INTEGER NOT NULL REFERENCES comic_storyboard(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES comic_role(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_comic_role_storyboard_storyboard_id ON comic_role_storyboard(storyboard_id);
CREATE INDEX idx_comic_role_storyboard_role_id ON comic_role_storyboard(role_id);
CREATE UNIQUE INDEX idx_comic_role_storyboard_unique ON comic_role_storyboard(storyboard_id, role_id);
```

## 3. 数据模型定义

### 3.1 Go 结构体定义

#### 3.1.1 动漫模型
```go
type Comic struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Title       string    `json:"title" gorm:"size:255;not null"`
    Brief       string    `json:"brief" gorm:"type:text"`
    Icon        string    `json:"icon" gorm:"size:500"`
    Bg          string    `json:"bg" gorm:"size:500"`
    UserPrompt  string    `json:"user_prompt" gorm:"type:text"`
    Status      string    `json:"status" gorm:"size:20;default:'pending'"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
    
    // 关联关系
    Roles       []ComicRole       `json:"roles,omitempty" gorm:"foreignKey:ComicID"`
    Sections    []ComicSection    `json:"sections,omitempty" gorm:"foreignKey:ComicID"`
}

func (Comic) TableName() string {
    return "comic"
}
```

#### 3.1.2 角色模型
```go
type ComicRole struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    ComicID   uint      `json:"comic_id" gorm:"not null"`
    Name      string    `json:"name" gorm:"size:100;not null"`
    ImageURL  string    `json:"image_url" gorm:"size:500"`
    Brief     string    `json:"brief" gorm:"type:text"`
    Voice     string    `json:"voice" gorm:"size:100"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 关联关系
    Comic     Comic     `json:"comic,omitempty" gorm:"foreignKey:ComicID"`
}

func (ComicRole) TableName() string {
    return "comic_role"
}
```

#### 3.1.3 章节模型
```go
type ComicSection struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    ComicID   uint      `json:"comic_id" gorm:"not null"`
    Index     int       `json:"index" gorm:"not null"`
    Detail    string    `json:"detail" gorm:"type:text;not null"`
    Status    string    `json:"status" gorm:"size:20;default:'pending'"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 关联关系
    Comic      Comic         `json:"comic,omitempty" gorm:"foreignKey:ComicID"`
    Storyboards []ComicStoryboard `json:"storyboards,omitempty" gorm:"foreignKey:SectionID"`
}

func (ComicSection) TableName() string {
    return "comic_section"
}
```

#### 3.1.4 运镜模型
```go
type ComicStoryboard struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    SectionID   uint      `json:"section_id" gorm:"not null"`
    ImagePrompt string    `json:"image_prompt" gorm:"type:text"`
    ImageURL    string    `json:"image_url" gorm:"size:500"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // 关联关系
    Section     ComicSection              `json:"section,omitempty" gorm:"foreignKey:SectionID"`
    Details     []ComicStoryboardDetail   `json:"details,omitempty" gorm:"foreignKey:StoryboardID"`
}

func (ComicStoryboard) TableName() string {
    return "comic_storyboard"
}
```

#### 3.1.5 运镜详情模型
```go
type ComicStoryboardDetail struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    StoryboardID uint      `json:"storyboard_id" gorm:"not null"`
    Detail       string    `json:"detail" gorm:"type:text"`
    RoleID       *uint     `json:"role_id"`
    TTSUrl       string    `json:"tts_url" gorm:"size:500"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    
    // 关联关系
    Storyboard   ComicStoryboard `json:"storyboard,omitempty" gorm:"foreignKey:StoryboardID"`
    Role         *ComicRole      `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

func (ComicStoryboardDetail) TableName() string {
    return "comic_storyboard_detail"
}
```

## 4. 数据库操作

### 4.1 基础CRUD操作
```go
// 创建漫画
func (r *ComicRepository) Create(comic *Comic) error {
    return r.db.Create(comic).Error
}

// 获取漫画列表
func (r *ComicRepository) GetList(limit, offset int) ([]Comic, error) {
    var comics []Comic
    err := r.db.Limit(limit).Offset(offset).Find(&comics).Error
    return comics, err
}

// 获取漫画详情
func (r *ComicRepository) GetByID(id uint) (*Comic, error) {
    var comic Comic
    err := r.db.Preload("Roles").Preload("Sections").First(&comic, id).Error
    return &comic, err
}

// 更新漫画状态
func (r *ComicRepository) UpdateStatus(id uint, status string) error {
    return r.db.Model(&Comic{}).Where("id = ?", id).Update("status", status).Error
}
```

### 4.2 复杂查询操作
```go
// 获取漫画的完整信息
func (r *ComicRepository) GetComicWithDetails(id uint) (*Comic, error) {
    var comic Comic
    err := r.db.Preload("Roles").
        Preload("Sections.Storyboards.Details.Role").
        First(&comic, id).Error
    return &comic, err
}

// 获取章节的分镜信息
func (r *ComicRepository) GetSectionStoryboards(sectionID uint) ([]ComicStoryboard, error) {
    var storyboards []ComicStoryboard
    err := r.db.Preload("Details.Role").
        Where("section_id = ?", sectionID).
        Find(&storyboards).Error
    return storyboards, err
}
```

## 5. 数据迁移

### 5.1 初始化迁移
```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Comic{},
        &ComicRole{},
        &ComicSection{},
        &ComicStoryboard{},
        &ComicStoryboardDetail{},
    )
}
```

### 5.2 数据种子
```go
func SeedData(db *gorm.DB) error {
    // 创建示例数据
    comic := &Comic{
        Title: "示例动漫",
        Brief: "这是一个示例动漫",
        Status: "completed",
    }
    
    if err := db.Create(comic).Error; err != nil {
        return err
    }
    
    return nil
}
```

## 6. 性能优化

### 6.1 索引优化
- 基于查询模式创建复合索引
- 定期分析查询性能
- 使用EXPLAIN分析查询计划

### 6.2 查询优化
- 使用预加载减少N+1查询
- 合理使用分页
- 避免SELECT *查询

### 6.3 连接池配置
```go
func SetupDatabase() *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        panic(err)
    }
    
    // 连接池配置
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db
}
```
