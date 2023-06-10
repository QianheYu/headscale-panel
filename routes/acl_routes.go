package routes

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
)

// InitAccessControlRoutes register access control routes
func InitAccessControlRoutes(r *gin.RouterGroup) gin.IRoutes {
	aclc := controller.NewAccessControlController()
	r.GET("/acl", aclc.GetAccessControl)
	r.POST("/acl", aclc.SetAccessControl)
	return r
}
