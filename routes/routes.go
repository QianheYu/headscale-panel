package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"headscale-panel/config"
	"headscale-panel/controller"
	"headscale-panel/log"
	"headscale-panel/middleware"
	"time"
)

// Init
func InitRoutes() *gin.Engine {
	// Set Nide
	gin.SetMode(config.Conf.System.Mode)

	// Create a route with default middleware:
	// Loggin and Recovery middleware
	r := gin.Default()
	// Create a route without middleware:
	// r := gin.New()
	// r.Use(gin.Recovery())

	// Enable rate limiting middleware
	// By default, fill one token every 50 milliseconds, up to a maximum of 200
	fillInterval := time.Duration(config.Conf.RateLimit.FillInterval)
	capacity := config.Conf.RateLimit.Capacity
	r.Use(middleware.RateLimitMiddleware(time.Millisecond*fillInterval, capacity))

	// Enable global CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Enable operation log middleware
	r.Use(middleware.OperationLogMiddleware())

	// Initialize JWT authentication middleware
	authMiddleware, err := middleware.InitAuth()
	if err != nil {
		log.Log.Panicf("Initialization of JWT middleware failed：%v", err)
		panic(fmt.Sprintf("Initialization of JWT middleware failed：%v", err))
	}

	r.GET("/.well-known/openid-configuration", controller.GetOpenIDConfiguration)

	// Route grouping
	apiGroup := r.Group("/" + config.Conf.System.UrlPathPrefix)
	InitOIDCRoutes(apiGroup, authMiddleware) // Register OIDC interface routes without authentication middleware

	// Register routes
	InitBaseRoutes(apiGroup, authMiddleware)         // Register basic routes, no need for JWT authentication middleware, no need for Casbin middleware
	InitUserRoutes(apiGroup, authMiddleware)         // Register user routes, require JWT authentication middleware, require Casbin authentication middleware
	InitRoleRoutes(apiGroup, authMiddleware)         // Register role routes, require JWT authentication middleware, require Casbin authentication middleware
	InitMenuRoutes(apiGroup, authMiddleware)         // Register menu routes, require JWT authentication middleware, require Casbin authentication middleware
	InitApiRoutes(apiGroup, authMiddleware)          // Register API routes, require JWT authentication middleware, require Casbin authentication middleware
	InitOperationLogRoutes(apiGroup, authMiddleware) // Register operation log routes, require JWT authentication middleware, require Casbin authentication middleware

	InitSystemRoutes(apiGroup, authMiddleware)  // Register system routes, require JWT authentication middleware, require Casbin authentication middleware
	InitConsoleRoutes(apiGroup, authMiddleware) // Register console routes, require JWT authentication middleware, require Casbin authentication middleware
	//InitNoticeRoutes(apiGroup, authMiddleware)  // SSE routes, require JWT authentication middleware, require Casbin authentication middleware
	//InitMessageRoutes(apiGroup, authMiddleware) // Message routes, require JWT authentication middleware, require Casbin authentication middleware

	log.Log.Info("Initialization route is completed！")
	return r
}
