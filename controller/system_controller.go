package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"headscale-panel/common"
	"headscale-panel/dto"
	"headscale-panel/log"
	"headscale-panel/response"
	task "headscale-panel/tasks"
	"headscale-panel/version"
	"headscale-panel/vo"
	"sync/atomic"
)

var flag atomic.Bool

type SystemController interface {
	GetInfo(c *gin.Context)
	GetStatus(c *gin.Context)
	Install(c *gin.Context)
}

type systemContorller struct {
	repo task.HeadscaleService
}

func NewSystemController() SystemController {
	return &systemContorller{}
}

func (s *systemContorller) GetInfo(c *gin.Context) {
	response.Success(c, &dto.SystemInfo{
		Version:   version.Version,
		BuildTime: version.BuildTime,
		Branch:    version.Branch,
		OS:        version.OS,
		Arch:      version.Arch,
		GoVersion: version.BuildGoVersion,
	}, "")
}

func (s *systemContorller) GetStatus(c *gin.Context) {
	response.Success(c, task.GetStatus(), "")
}

func (s *systemContorller) Install(c *gin.Context) {
	req := &vo.InstallRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, nil, "param error")
		log.Log.Errorf("param error: %v", err)
		return
	}

	if err := common.Validate.Struct(req); err != nil {
		response.Fail(c, nil, "param error")
		log.Log.Errorf("param error: %v", err)
		return
	}

	if flag.Load() {
		response.Success(c, nil, "system is installing, please wait")
		return
	}

	flag.Store(true)
	defer func() {
		flag.Store(false)
	}()
	switch req.State {
	case "version":
		releases, err := s.repo.GetVersions()
		if err != nil {
			response.Fail(c, nil, "get versions error")
			return
		}
		response.Success(c, releases, "")
	case "install":
		log.Log.Debugf("get install request, id is %d\n", req.ID)
		if err := s.repo.Install(req.ID); err != nil {
			response.Fail(c, nil, "install headscale error")
			return
		}
		s.repo.Stop(context.Background())
		if err := s.repo.Start(); err != nil {
			response.Fail(c, nil, err.Error())
			return
		}
		response.Success(c, nil, "install headscale success")
	case "upgrade":
		go func() {
			if err := s.repo.Update(); err != nil {
				log.Log.Errorf("update error: %s", err)
				return
			}
			if err := s.repo.Stop(context.Background()); err != nil {
				log.Log.Errorf("stop headscale error: %v", err)
			}
			if err := s.repo.Start(); err != nil {
				log.Log.Errorf("start headscale error: %v", err)
				return
			}
		}()
		response.Success(c, nil, "headscale updated")
	default:
		response.Fail(c, nil, "unknown state")
	}
}
