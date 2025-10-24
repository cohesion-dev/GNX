# 前端代码结构设计

```
/www/
├── package.json
├── next.config.js
├── tailwind.config.js
├── tsconfig.json
├── public/
│   ├── images/
│   ├── icons/
│   └── favicon.ico
├── src/
│   ├── pages/                    # 页面目录
│   │   ├── _app.tsx             # 应用入口
│   │   ├── _document.tsx        # 文档配置
│   │   ├── index.tsx            # 首页
│   │   ├── comic/               # 漫画相关页面
│   │   │   ├── index.tsx        # 漫画列表页
│   │   │   ├── add.tsx          # 新增漫画页
│   │   │   └── detail/
│   │   │       ├── [id].tsx     # 漫画详情页
│   │   │       ├── [id]/read.tsx # 漫画播放页
│   │   │       └── [id]/section/
│   │   │           └── add.tsx   # 新增章节页
│   │   └── api/                 # API路由（如果需要）
│   ├── components/              # 组件目录
│   │   ├── common/             # 通用组件
│   │   │   ├── Layout.tsx       # 布局组件
│   │   │   ├── Header.tsx       # 头部组件
│   │   │   ├── Footer.tsx       # 底部组件
│   │   │   ├── Loading.tsx      # 加载组件
│   │   │   ├── ErrorBoundary.tsx # 错误边界
│   │   │   └── Modal.tsx        # 模态框组件
│   │   ├── comic/              # 漫画相关组件
│   │   │   ├── ComicList.tsx    # 漫画列表组件
│   │   │   ├── ComicCard.tsx    # 漫画卡片组件
│   │   │   ├── ComicDetail.tsx  # 漫画详情组件
│   │   │   ├── ComicForm.tsx    # 漫画表单组件
│   │   │   ├── ComicPlayer.tsx  # 漫画播放器组件
│   │   │   ├── RoleList.tsx     # 角色列表组件
│   │   │   ├── RoleCard.tsx     # 角色卡片组件
│   │   │   ├── SectionList.tsx  # 章节列表组件
│   │   │   ├── SectionCard.tsx  # 章节卡片组件
│   │   │   └── StoryboardViewer.tsx # 分镜查看器组件
│   │   └── forms/              # 表单组件
│   │       ├── FileUpload.tsx   # 文件上传组件
│   │       ├── TextInput.tsx    # 文本输入组件
│   │       └── TextArea.tsx     # 文本域组件
│   ├── stores/                 # MobX状态管理
│   │   ├── index.ts            # Store入口
│   │   ├── ComicStore.ts       # 漫画状态管理
│   │   ├── PlayerStore.ts      # 播放器状态管理
│   │   └── UIStore.ts          # UI状态管理
│   ├── services/               # 服务层
│   │   ├── api.ts              # API客户端
│   │   ├── comicService.ts     # 漫画服务
│   │   ├── playerService.ts    # 播放器服务
│   │   └── fileService.ts       # 文件服务
│   ├── types/                  # 类型定义
│   │   ├── index.ts            # 类型导出
│   │   ├── comic.ts            # 漫画类型
│   │   ├── player.ts           # 播放器类型
│   │   └── api.ts              # API类型
│   ├── hooks/                  # 自定义Hooks
│   │   ├── useComic.ts         # 漫画相关Hook
│   │   ├── usePlayer.ts        # 播放器相关Hook
│   │   ├── useFileUpload.ts    # 文件上传Hook
│   │   └── useWebSocket.ts     # WebSocket Hook
│   ├── utils/                  # 工具函数
│   │   ├── index.ts            # 工具函数导出
│   │   ├── format.ts           # 格式化工具
│   │   ├── validation.ts       # 验证工具
│   │   ├── storage.ts          # 存储工具
│   │   └── constants.ts         # 常量定义
│   └── styles/                 # 样式文件
│       ├── globals.css         # 全局样式
│       └── components.css       # 组件样式
└── README.md
```