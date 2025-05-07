package api

import (
	"invite-code-service/pkg/config"
	"invite-code-service/pkg/db"
	"net/http"

	_ "invite-code-service/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouters(db *db.WrapDb, cfg *config.ConfigApi) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/static", "./static")
	router.Use(Cors())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handler := NewHandler(db, cfg)
	router.GET("/api/v1/invite/tasks", handler.GetTasks)
	router.GET("/api/v1/invite/userStatus", handler.GetUserStatus)

	router.POST("/api/v1/invite/bind", handler.HandlePostBind)

	return router
}
