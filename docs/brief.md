我现在要做一个智能动漫生成系统：

前端设计：
### 项目架构

- 前端根目录：`/www`
- Next.js + TypeScript + Mobx + Tailwind
  - Next.js 全部使用 CSR 模式，不使用 SSR 模式
- 需要考虑移动端适配

### 项目目录结构

- `/www/src/pages/xxx`: 页面根目录
- `/www/src/apis/xxx`: 接口根目录

### MVP 需求拆解

1. 漫画列表页面（目页面路由：`/comic`）
  - 漫画列表模块：展示每个漫画的名称、封面、状态信息，以及跳转「漫画详情页面」入口
    - 状态包括「完成」和「生成中」
    - 后端接口：查询漫画列表接口

2. 新增漫画页面（页面路由：`/comic/add`）
  - 新增漫画模块：选择本地小说文件，以及输入漫画名字、文本提示词，一起提交后后端生成漫画
    - 生成任务初始化完成后，跳转「漫画播放页面」
    - 后端接口：新增漫画接口

3. 漫画详情页面（页面路由：`/comic/detail/[id]`）
  - 漫画概览模块：展示漫画名称、封面、简介
    - 后端接口：查询漫画概览接口
  - 漫画角色列表模块：展示每个角色的头像、名字、简介
    - 后端接口：查询漫画概览接口（包含角色信息）
  - 漫画章节列表模块：展示章节名称、第几章、是否已阅读过
    - 后端接口：查询漫画章节列表

4. 新增漫画章节页面（页面路由：`/comic/detail/[id]/section/add`）
  - 新增漫画章节模块：输入最新章节文本，提交后后端给已有的漫画生成新的漫画章节
    - 生成任务初始化完成后，跳转「漫画播放页面」并跳转到对应章节
    - 后端接口：新增章节接口

5. 漫画播放页面（页面路由：`/comic/detail/[id]/read`）：
  - 漫画播放主模块：展示章节内容、分镜图片、音频播放
    - 后端接口：获取章节内容接口、获取章节分镜接口
  - 漫画顶部导航模块：展示章节名称、第几章，以及返回按钮
    - 返回按钮跳转到「漫画详情页面」
  - 漫画播放控制底部模块：展示章节名称、第几章、是否已阅读过
    - 后端接口：查询漫画章节列表
    - 子模块：
      - 漫画章节列表子模块：展示章节名称、第几章、是否已阅读过
        - 后端接口：查询漫画章节列表
      - 漫画章节进度条子模块：展示章节播放进度，以及支持切换播放进度
      - 漫画播放设置子模块：播放速度、音量、自动播放等设置



后端设计：
开发语言：Go + gogin + gorm 
数据库： postgrasql
图片和音频存储：七牛 storage

### API 接口设计

1. 漫画相关接口：
   - GET /api/comic - 查询漫画列表
   - POST /api/comic - 新增漫画
   - GET /api/comic/{id} - 查询漫画详情
   - GET /api/comic/{id}/sections - 查询漫画章节列表

2. 章节相关接口：
   - POST /api/comic/{id}/section - 新增章节
   - GET /api/comic/{id}/section/{section_id}/content - 获取章节内容
   - GET /api/comic/{id}/section/{section_id}/storyboards - 获取章节分镜
   - GET /api/comic/{id}/section/{section_id}/storyboard/{storyboard_id}/image - 获取章节图片

3. TTS 播放接口
   - GET /api/tts/{storyboard_tts_id} - 获取章节 tts 播放音频流

数据库：
1. 动漫表：comic
    - id:
    - title: 动漫名（前端传入）
    - brief: 动漫简介 （AI 生成）
    - icon: 动漫 icon （AI 异步生成）
    - bg: 动漫背景图（AI 异步生成）
    - user_prompt: 用户提示词（前端传入）
    - status: AI 生成状态（pending: 生成中, completed: 已完成, failed: 生成失败）

2. 动漫角色表：comic_role
    - id:
    - comic_id: 动漫 ID（索引）
    - name: 角色名称（AI 生成）
    - image_url: 角色图片（AI 生成）
    - brief: 简介（AI 生成）
    - voice: 音色（AI 生成）

3. 动漫章节表：comic_section
    - id:
    - comic_id: 动漫 ID（索引）
    - index: 章节数
    - detail: 章节原文
    - status: AI 生成状态（pending: 生成中, completed: 已完成, failed: 生成失败）其下所有的 storyboard 的 image_url 都有值时即为 completed

4. 动漫运镜表：comic_storyboard 
    - id:
    - section_id: 章节 ID（索引）
    - image_prompt: 运镜提示词
    - image_url: 运镜图片链接

5. 动漫运镜详情表：comic_storyboard_detail
    - id:
    - storyboard_id: 运镜表 ID（索引）
    - detail: 原文
    - role_id: 角色 ID（索引）
    - tts_url: tts 音频链接

6. 动漫角色运镜关系表：comic_role_storyboard
    - id:
    - storyboard_id: 运镜表 ID（索引）
    - role_id: 角色 ID（索引）




动漫生成流程：
1. 让 AI 大模型归纳出下面信息：
    - 小说的简介
    - 小数的主要角色 数组
          -  [{
            角色名称,
            简介, 
            音色,
            Range: [1,2,4,9],
          }]

2. 使用 AI 异步生成小说的 icon 和 背景大图

3. 按章节为单位切分分镜内容
[
{ // 分镜
    details: [
        {
            voice， // 可选
            detail, // 可选
            角色名称, // 可选
        }
    ],
    image_prompt,
},
...
]

4. 使用 AI 异步生成分镜图片




AI 模型交互:
模型: OpenAI 的接口，文生文 文生图 图生图 
TTS：七牛的 tts
1. AIMoiveGen 
    - Section(sec, {baseRoles, desc}) {newRoles, desc, items}
    - for each section:
        Section(section, )

2. GenImage(baseImages, prompt) image

3. TTS(音色，content) AudioStream




