package routes

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
)

// InitNodesRoutes register machine routes about get, modify, delete and etc.
func InitNodesRoutes(r *gin.RouterGroup) gin.IRoutes {
	nodes := controller.NewNodesController()
	r.GET("/machine", nodes.GetNodes)
	r.POST("/machine", nodes.StateNodes)
	r.DELETE("/machine", nodes.DeleteNode)
	r.PUT("/machine", nodes.MoveNode)
	r.PATCH("/machine", nodes.SetTags)
	return r
}
