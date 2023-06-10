package routes

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
)

// InitPreAuthKey register routes about user management their PreAuthKey
func InitPreAuthKey(r *gin.RouterGroup) gin.IRoutes {
	preAuthKeyController := controller.NewPreAuthKeyController()
	r.GET("/preauthkey", preAuthKeyController.ListPreAuthKey)
	r.POST("/preauthkey", preAuthKeyController.CreatePreAuthKey)
	r.DELETE("/preauthkey", preAuthKeyController.ExpirePreAuthKey)
	return r
}
