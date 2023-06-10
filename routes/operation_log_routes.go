package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
	"headscale-panel/middleware"
)

func InitOperationLogRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	operationLogController := controller.NewOperationLogController()
	router := r.Group("/log")
	// Enable JWT authentication middleware
	router.Use(authMiddleware.MiddlewareFunc())
	// Enable Casbin authentication middleware
	router.Use(middleware.CasbinMiddleware())
	{
		router.GET("/operation/list", operationLogController.GetOperationLogs)
		router.DELETE("/operation/delete/batch", operationLogController.BatchDeleteOperationLogByIds)
	}
	return r
}
