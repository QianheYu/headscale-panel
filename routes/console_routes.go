package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/middleware"
)

// InitConsoleRoutes register console and sub routes and them must use authentication middleware
func InitConsoleRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	consoleGroup := r.Group("/console")
	// Enable JWT authentication middleware
	consoleGroup.Use(authMiddleware.MiddlewareFunc())
	// Enable Casbin authentication middleware
	consoleGroup.Use(middleware.CasbinMiddleware())

	InitPreAuthKey(consoleGroup)          // Register PreAuthKey API
	InitRouteRoutes(consoleGroup)         // Register Route API
	InitNodesRoutes(consoleGroup)         // Register Machine API
	InitAccessControlRoutes(consoleGroup) // Register ACL API
}
