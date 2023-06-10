package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
	"headscale-panel/middleware"
)

func InitSystemRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	s := r.Group("/system")
	// Enable JWT authentication middleware
	s.Use(authMiddleware.MiddlewareFunc())
	// Enable Casbin authentication middleware
	s.Use(middleware.CasbinMiddleware())

	headscale := controller.NewHeadscaleConfigController()
	s.GET("/headscale", headscale.GetHeadscaleConfig)
	s.POST("/headscale", headscale.SetHeadscaleConfig)
	// 功能预留
	//s.POST("/headscale/upload/:target", headscale.Upload)

	system := controller.NewSystemController()
	s.GET("/info", system.GetInfo)
	s.GET("/status", system.GetStatus)
	s.POST("/install", system.Install)
	return r
}
