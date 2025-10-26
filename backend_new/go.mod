module github.com/cohesion-dev/GNX/backend_new

go 1.22

require (
	github.com/cohesion-dev/GNX/ai v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/qiniu/go-sdk/v7 v7.21.1
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.12
)

replace github.com/cohesion-dev/GNX/ai => ../ai
