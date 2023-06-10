package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
)

type IMachinesController interface {
	GetMachines(c *gin.Context)   // method: get
	StateMachines(c *gin.Context) // method: post 重新构建request结构
	DeleteMachine(c *gin.Context) // method: delete
	MoveMachine(c *gin.Context)
	SetTags(c *gin.Context)
}

type machinesController struct {
	userRepo     repository.IUserRepository
	machinesRepo repository.HeadscaleMachinesRepository
}

func NewMachinesController() IMachinesController {
	return &machinesController{userRepo: repository.NewUserRepository(), machinesRepo: repository.NewMachinesRepo()}
}

// GetMachines
func (m *machinesController) GetMachines(c *gin.Context) {
	//user, err := m.userRepo.GetCurrentUser(c)
	//if err != nil {
	//	response.Fail(c, nil, "Failed to get machine")
	//	log.Log.Errorf("get current user error: %v", err)
	//	return
	//}
	machines, err := m.machinesRepo.ListMachinesWithUser("")
	if err != nil && err.Error() != "rpc error: code = Unknown desc = User not found" {
		response.Fail(c, nil, "Failed to get machines")
		log.Log.Errorf("get machine error: %v", err)
		return
	}
	response.Success(c, machines, "success")
}

// StateMachines 对设备进行注册、过期、重命名操作
func (m *machinesController) StateMachines(c *gin.Context) {
	req := &vo.EditMachineRequest{}
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	var err error
	var data interface{}
	switch req.State {
	case "rename":
		// rename machine
		data, err = m.machinesRepo.RenameMachineWithNewName(req.MachineId, req.Name)
	case "expire":
		// expire machine
		data, err = m.machinesRepo.ExpireMachineWithId(req.MachineId)
	case "register":
		// register machine
		var user model.User
		user, err = m.userRepo.GetCurrentUser(c)
		if err != nil {
			break
		}
		data, err = m.machinesRepo.RegisterMachineWithKey(user.Name, req.Nodekey)
	default:
		response.Fail(c, nil, "params error")
		return
	}
	if err != nil {
		response.Fail(c, nil, "Failed to operate")
		log.Log.Errorf("operate machine error: %v", err)
		return
	}
	response.Success(c, data, "success")
}

// MoveMachine move machine to another user
func (m *machinesController) MoveMachine(c *gin.Context) {
	req := &vo.MoveMachineRequest{}

	// Bind parameters
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	machine, err := m.machinesRepo.MoveMachine(req)
	if err != nil {
		response.Fail(c, nil, "Failed to move machine")
		log.Log.Errorf("move machine error: %v", err)
		return
	}
	response.Success(c, machine, "move success")
}

// DeleteMachine delete machine
func (m *machinesController) DeleteMachine(c *gin.Context) {
	req := &vo.DeleteMachineRequest{}

	// Bind parameters
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	if err := m.machinesRepo.DeleteMachine(req); err != nil {
		response.Fail(c, nil, "Failed to delete machine")
		log.Log.Errorf("delete machine error: %v", err)
		return
	}
	response.Success(c, nil, "success")
}

// SetTags set tag on machine
func (m *machinesController) SetTags(c *gin.Context) {
	req := &vo.SetTagsRequest{}

	// Bind parameters
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	data, err := m.machinesRepo.SetTagsWithStringSlice(req.MachineId, req.Tags)
	if err != nil {
		response.Fail(c, nil, "Failed to set tag")
		log.Log.Errorf("set tag error: %v", err)
		return
	}
	response.Success(c, data, "set tags success")
}
