/* This part of the code is reserved for functionality */
package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
	"headscale-panel/middleware"
)

func InitNoticeRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	notice := controller.NewNoticeController()
	// Enable JWT authentication middleware
	r.Use(authMiddleware.MiddlewareFunc())
	// Enable Casbin authentication middleware
	r.Use(middleware.CasbinMiddleware())

	r.GET("/notice", notice.Controller)
	return r
}
