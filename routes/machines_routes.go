package routes

import (
	"github.com/gin-gonic/gin"
	"headscale-panel/controller"
)

// InitMachinesRoutes register machine routes about get, modify, delete and etc.
func InitMachinesRoutes(r *gin.RouterGroup) gin.IRoutes {
	machines := controller.NewMachinesController()
	r.GET("/machine", machines.GetMachines)
	r.POST("/machine", machines.StateMachines)
	r.DELETE("/machine", machines.DeleteMachine)
	r.PUT("/machine", machines.MoveMachine)
	r.PATCH("/machine", machines.SetTags)
	return r
}
