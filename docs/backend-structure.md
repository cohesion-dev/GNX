# 后端代码结构设计

```
/backend/
├── go.mod
├── go.sum
├── main.go
├── config/
│   ├── config.go
│   └── database.go
├── internal/
│   ├── app/
│   │   ├── server.go
│   │   └── routes.go
│   ├── handlers/
│   │   ├── comic.go
│   │   ├── section.go
│   │   ├── player.go
│   │   └── tts.go
│   ├── services/
│   │   ├── comic_service.go
│   │   ├── ai_service.go
│   │   ├── storage_service.go
│   │   └── task_service.go
│   ├── repositories/
│   │   ├── comic_repository.go
│   │   ├── section_repository.go
│   │   └── storyboard_repository.go
│   ├── models/
│   │   ├── comic.go
│   │   ├── section.go
│   │   ├── storyboard.go
│   │   └── role.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logging.go
│   │   └── recovery.go
│   ├── utils/
│   │   ├── response.go
│   │   ├── validation.go
│   │   ├── file.go
│   │   └── constants.go
│   └── tasks/
│       ├── comic_generation.go
│       ├── image_generation.go
│       └── tts_generation.go
├── pkg/
│   ├── ai/
│   │   ├── openai.go
│   │   └── qiniu_tts.go
│   └── storage/
│       └── qiniu.go
├── migrations/
│   ├── 001_create_comic_table.sql
│   ├── 002_create_role_table.sql
│   └── 003_create_section_table.sql
├── scripts/
│   ├── migrate.go
│   └── seed.go
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── docs/
│   ├── api.md
│   └── deployment.md
└── tests/
    ├── handlers/
    ├── services/
    └── repositories/
```
