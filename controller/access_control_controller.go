package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"headscale-panel/common"
	"headscale-panel/repository"
	"headscale-panel/response"
	task "headscale-panel/tasks"
	"headscale-panel/vo"
)

type AccessControlController interface {
	GetAccessControl(c *gin.Context)
	SetAccessControl(c *gin.Context)
}

type accessControl struct {
	repo repository.AccessControlRepository
	h    task.HeadscaleService
}

func NewAccessControlController() AccessControlController {
	return &accessControl{repo: repository.NewAccessControlRepository()}
}

// Get the contents of the ACL file, available for standalone deployments
func (a *accessControl) GetAccessControl(c *gin.Context) {
	data, err := a.repo.GetAccessControl()
	if err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	response.Success(c, data, "success")
}

// Set the content of the ACL file, available for stand-alone deployment
func (a *accessControl) SetAccessControl(c *gin.Context) {
	req := &vo.SetAccessControlRequest{}
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

	if err := a.repo.SetAccessControl(req.Content); err != nil {
		response.Fail(c, nil, err.Error())
		return
	}

	// Restart headscale
	a.h.Stop(context.Background())
	a.h.Start()
	response.Success(c, nil, "save success")
}
