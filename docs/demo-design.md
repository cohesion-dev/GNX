## Demo 漫画生成设计思路（重构版）

### 目标
- 从本地小说文本生成分章节的漫画素材：分镜图片、面板音频、角色概念图与清单。
- 保持角色外观与风格一致性，支持并发生成与失败回退，保证输出目录结构与原 Demo 一致。

### 输入参数
- `-input`: 小说文本文件路径（必填）
- `-output`: 输出目录（默认 `output`）
- `-title`: 小说标题（默认 `未知小说`）
- `-max-chapters`: 最大处理章节数（0 表示全部）
- `-image-style`: 图像风格提示词前缀（默认 `卡通风格，`）

### 总体流程
1. 章节切分：基于 `SplitChaptersFromFile` 将小说文本按章节拆分。
2. 章节摘要：调用 `SummaryChapter` 输出角色特征、可用音色、分镜页与面板结构。
3. 角色概念图：
   - 维护全局 `characterRegistry` 与 `characterAssets` 用于跨章节稳定角色一致性。
   - 拼接风格前缀与角色概念图提示词，生成或复用概念图；写入章节与全局 `manifest.json`。
4. 分镜图片：
   - 按页并发生成，优先使用对应角色概念图做参考（多图/单图 img2img），失败回退到 text2img。
5. 面板音频：
   - 每页内并发遍历面板与文本段，调用 TTS 生成音频文件。
6. 清单与产物：
   - 每章输出 `storyboard.json`、页图片与面板音频；角色清单写入章节与全局目录。

### 关键策略
- 角色一致性：`normalizeCharacterKey` 与 `sanitizeCharacterFileStem` 实现稳定命名与检索；概念图复用与最小化重生成。
- 提示词策略：将 `-image-style` 前缀与场景/角色提示词拼接，保证风格统一同时保留细节。
- 并发与日志：页级并发生成图片，面板级并发生成音频；通过互斥锁串联日志输出，避免乱序。
- 降级回退：img2img 失败时自动回退到 text2img；多图参考失败回退到单图或纯文本。

### 目录结构（示例）
- `output/characters/`：全局角色概念图与提示词
- `output/chapter_001/`：章节 1 图像、音频与角色清单
  - `characters/manifest.json` / 概念图与提示词副本
  - `storyboard.json` / 分镜结构与角色特征
  - `page_001.png` / 等
  - `page_001_panel_01_audio_001.mp3` / 等

### 模块划分（重构）
- `types.go`：数据结构（角色资产、清单条目与清单）
- `utils.go`：工具函数（键规范化、角色有序特征、文件复制、file stem 生成、分镜角色引用收集）
- `main.go`：CLI 入口与流程编排（调用 gnxaigc 服务生成图像与音频）

### 后续演进建议
- 抽象 `characters.go` 与 `storyboard.go` 以分离职责，便于单元测试与替换实现。
- 增加生成缓存与断点恢复，避免重复调用 AIGC 服务。

