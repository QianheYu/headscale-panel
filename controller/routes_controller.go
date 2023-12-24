package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
	"strconv"
)

type RouteController interface {
	GetMachinesRoute(c *gin.Context) // method: get
	DeleteRoute(c *gin.Context)      // method: delete
	SwitchRoute(c *gin.Context)      // method: post
}

type routeController struct {
	repo repository.HeadscaleRouteRepository
}

func NewRoutesController() RouteController {
	return &routeController{repo: repository.NewRouteRepo()}
}

func (r *routeController) GetMachinesRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.DefaultQuery("machine_id", "0"), 10, 64)
	if err != nil {
		log.Log.Errorf("parse int error: %w", err)
		response.Fail(c, nil, "param error")
		return
	}

	var routes interface{}
	if id == 0 {
		routes, err = r.repo.GetRoutes()
	} else {
		routes, err = r.repo.GetNodeRoutesWithId(id)
	}
	//routes, err := r.repo.GetMachineRoutesWithId(id)
	if err != nil {
		response.Fail(c, nil, "Get routes error")
		return
	}
	response.Success(c, routes, "success")
}

func (r *routeController) DeleteRoute(c *gin.Context) {
	req := &vo.DeleteRouteRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// validate data
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	if err := r.repo.DeleteRoute(req); err != nil {
		response.Fail(c, nil, "Failed to delete route")
		log.Log.Errorf("delete route error: %v", err)
		return
	}
	response.Success(c, nil, "success")
}

func (r *routeController) SwitchRoute(c *gin.Context) {
	req := &vo.SwitchRouteRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// validate data
	if err := common.Validate.Struct(req); err != nil {
		log.Log.Errorf("param error: %v", err.(validator.ValidationErrors)[0].Translate(common.Trans))
		response.Fail(c, nil, "param error")
		return
	}

	err := r.repo.SwitchRoute(req)
	if err != nil {
		response.Fail(c, nil, fmt.Sprintf("Failed to switch to %v", req.Enable))
		log.Log.Errorf("Failed to switch route to %v, %v", req.Enable, err)
		return
	}
	response.Success(c, nil, "success")
}
