## 错误响应示例：

```json5
{
  code: 400,
  message: "Bad Request",
  details: "Invalid request parameters",
}
```

## API 列表

### 获取用户的所有漫画列表

```text
GET /comics/
```

```json5
{
  code: 200,
  message: "成功",
  data: {
    comics: [
      {
        id: "string", // 漫画唯一标识符
        icon_image_id: "string", // 漫画封面图片ID
        background_image_id: "string", // 漫画背景图片ID
        title: "string", // 漫画标题
        status: "<failed|completed|pending>", // 漫画状态
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
    ],
    total: 100,
    page: 1,
    limit: 10,
  },
}
```

### 上传小说文件并创建新的漫画项目

```text
POST /comics/
```

```multipart
"title": "string", // 漫画标题
"user_prompt": "string", // 用户提示词
"file": "<file>", // 小说文件，支持txt、docx格式
```

返回

```json5
{
  code: 200,
  message: "成功",
  data: {
    id: "string", // 漫画唯一标识符
  },
}
```

### 获取漫画详情和章节列表

```text
GET /comics/{comic_id}/
```

```json5
{
  code: 200,
  message: "成功",
  data: {
    id: "string", // 漫画唯一标识符
    title: "string", // 漫画标题
    user_prompt: "string", // 用户提示词
    status: "<failed|completed|pending>", // 漫画状态
    roles: [
      // 漫画中的角色列表
      {
        name: "string", // 角色名称
        brief: "string", // 角色简介
        image_id: "string", // 角色形象图片ID
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
    ],
    sections: [
      {
        id: "string", // 章节唯一标识符
        title: "string", // 章节标题
        index: 1, // 章节索引
        status: "<failed|completed|pending>", // 章节状态
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      },
      ...
    ], // 漫画章节列表
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
}
```

### 创建新章节

```text
POST /comics/{comic_id}/sections/
```

```multipart
"title": "string", // 章节标题
"content": "string", // 章节内容
```

返回

```json5
{
  code: 200,
  message: "成功",
  data: {
    id: "string", // 章节唯一标识符
    index: 1, // 章节索引
  },
}
```

### 获取章节详情和页面列表

```text
GET /comics/{comic_id}/sections/{section_id}/
```

```json5
{
  code: 200,
  message: "成功",
  data: {
    id: "string", // 章节唯一标识符
    title: "string", // 章节标题
    index: 1, // 章节索引
    status: "<failed|completed|pending>", // 章节状态
    pages: [
      // 章节页面列表
      {
        id: "string", // 页面唯一标识符，同时也是获取图片的ID
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        details: [
          {
            id: "string", // 文字唯一标识，同时也是TTS唯一标识符，同时也是触发TTS生成的ID
            content: "string", // 页面文字内容
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        ],
      },
    ], // 章节页面列表
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
}
```

### 获取图片链接

```text
GET /images/{image_id}/url
```

```json5
{
  code: 200,
  message: "成功",
  data: {
    url: "string", // 图片链接，有效期1小时
  },
}
```

如果此时图片还未生成，返回 404 错误。

### 触发文字转语音生成

```text
GET /tts/{tts_id}
```

音频数据流直接返回，Content-Type 为 audio/\*
