package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
	"headscale-panel/middleware"
)

func InitOIDCRoutes(r *gin.RouterGroup, authorization *jwt.GinJWTMiddleware) gin.IRoutes {
	oidc := controller.NewOIDCController()

	r.POST("/oidc/token", oidc.Token)
	r.GET("/oidc/user_info", oidc.GetUserInfo)
	r.GET("/oidc/jwk", oidc.JWKs)
	r.POST("/oidc/authorize", authorization.MiddlewareFunc(), middleware.CasbinMiddleware(), oidc.Authorize)

	return r
}
