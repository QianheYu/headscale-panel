package routes

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
)

// InitRouteRoutes register route about machine accessed sub route
func InitRouteRoutes(r *gin.RouterGroup) gin.IRoutes {
	routes := controller.NewRoutesController()
	r.GET("/route", routes.GetMachinesRoute)
	r.PATCH("/route", routes.SwitchRoute)
	r.DELETE("/route", routes.DeleteRoute)
	return r
}
