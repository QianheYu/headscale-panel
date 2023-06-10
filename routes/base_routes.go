package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// Register basic routes
func InitBaseRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	router := r.Group("/base")
	{
		// No authentication required for login/logout token refresh
		router.POST("/login", authMiddleware.LoginHandler)
		router.POST("/logout", authMiddleware.LogoutHandler)
		router.POST("/refreshToken", authMiddleware.RefreshHandler)
	}
	return r
}
