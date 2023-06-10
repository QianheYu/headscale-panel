package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
	"headscale-panel/middleware"
)

func InitApiRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	apiController := controller.NewApiController()
	router := r.Group("/api")
	// Enable JWT authentication middleware
	router.Use(authMiddleware.MiddlewareFunc())
	// Enable Casbin authentication middleware
	router.Use(middleware.CasbinMiddleware())
	{
		router.GET("/list", apiController.GetApis)
		router.GET("/tree", apiController.GetApiTree)
		router.POST("/create", apiController.CreateApi)
		router.PATCH("/update/:apiId", apiController.UpdateApiById)
		router.DELETE("/delete/batch", apiController.BatchDeleteApiByIds)
	}

	return r
}
