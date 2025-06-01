package router

import (
	"github.com/gin-gonic/gin"
	"github.com/virhanali/filmnesia/api-gateway/internal/config"
	"github.com/virhanali/filmnesia/api-gateway/internal/handler"
)

func SetupRouter(cfg config.Config) *gin.Engine {
	router := gin.Default()

	userServiceProxy := handler.NewReverseProxy(cfg.UserServiceURL)
	userRoutes := router.Group("/api/v1/users")
	{
		userRoutes.Any("/*proxyPath", userServiceProxy)
	}

	router.GET("/gateway/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "API Gateway is healthy"})
	})

	return router
}
