package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/tasks"
	"headscale-panel/vo"
)

type headscaleConfigController struct {
	repo repository.HeadscaleConfigRepository
	h    task.HeadscaleService
}

type IHeadscaleConfigController interface {
	GetHeadscaleConfig(c *gin.Context)
	SetHeadscaleConfig(c *gin.Context)
	//Upload(c *gin.Context)
}

func NewHeadscaleConfigController() IHeadscaleConfigController {
	return &headscaleConfigController{repo: repository.NewHeadscaleConfigRepository()}
}

// GetHeadscaleConfig Get the content of the Headscale configuration file for standalone deployments,
// get the grpc host address and ApiKey from the database for multi-machine deployments
func (s *headscaleConfigController) GetHeadscaleConfig(c *gin.Context) {
	var err error
	var settings interface{}
	if config.GetMode() < config.MULTI {
		settings, err = s.repo.GetHeadscaleConfigFromFile(config.Conf.Headscale.Config)
	} else {
		settings, err = s.repo.GetHeadscaleConfigFromDB()
	}

	//settings, err := s.repo.GetHeadscaleConfig()
	if err != nil {
		response.Fail(c, nil, err.Error())
		return
	}
	response.Success(c, settings, "success")
}

// SetHeadscaleConfig Save the Headscale configuration file during stand-alone deployment,
// and save the host address and ApiKey of the grpc connection to the database during separate deployment
func (s *headscaleConfigController) SetHeadscaleConfig(c *gin.Context) {
	var req vo.SystemSettingHeadscale
	// Bind parameters
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Select the setting saving method according to the operation mode,
	// the stand-alone deployment mode saves the headscale configuration file,
	// and the remote call mode saves the grpc connection configuration to the database
	var err error
	if config.GetMode() < config.MULTI {
		err = s.repo.SetHeadscaleConfigFromFile(config.Conf.Headscale.Config, &req)
	} else {
		err = s.repo.SetHeadscaleConfigFromDB(&req)
	}

	if err != nil {
		response.Fail(c, nil, err.Error())
		return
	}

	// Restart related services
	go func() {
		// Determine whether to restart the headscale according to the operating mode,
		// and it will restart when deploying on a single machine
		if config.GetMode() < config.MULTI {
			s.h.Stop(context.Background())
			s.h.Start()
		}
		if err := task.HeadscaleControl.ReConnect(); err != nil {
			log.Log.Errorf("grpc err: %v", err)
		}
	}()

	response.Success(c, nil, "success")
}

// Function reserved
//func (s *headscaleConfigController) Upload(c *gin.Context) {
//	target := c.Param("target")
//
//	// Bind and Validate data
//	file, header, err := c.Request.FormFile(target)
//	if err != nil {
//		response.Fail(c, nil, "")
//		return
//	}
//	defer file.Close()
//	if header.Header.Get("Content-Type") != "multipart/form-data" {
//		response.Fail(c, nil, "Upload error")
//		return
//	}
//
//	switch target {
//	case "cert":
//		err = s.repo.SetHeadscaleCert(file)
//	case "key":
//		err = s.repo.SetHeadscaleKey(file)
//	case "ca":
//		err = s.repo.SetHeadscaleCA(file)
//	default:
//		response.Fail(c, nil, "Method error")
//		return
//	}
//	if err != nil {
//		response.Fail(c, nil, "Upload failed")
//		return
//	}
//	response.Success(c, nil, "")
//}
