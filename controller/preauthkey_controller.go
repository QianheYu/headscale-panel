package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
	"time"
)

type PreAuthKeyController interface {
	ListPreAuthKey(c *gin.Context)
	CreatePreAuthKey(c *gin.Context)
	ExpirePreAuthKey(c *gin.Context)
}

type preAuthKeyController struct {
	repo     repository.HeadscalePreAuthKeyRepository
	userRepo repository.IUserRepository
}

// NewPreAuthKeyController new a controller to
// Only obtain the user's own PreAuthKey and cannot operate other users' PreAuthKey
func NewPreAuthKeyController() PreAuthKeyController {
	return &preAuthKeyController{repo: repository.NewPreAuthkeyRepo(), userRepo: repository.NewUserRepository()}
}

// ListPreAuthKey get user PreAuthKey by current user
func (p *preAuthKeyController) ListPreAuthKey(c *gin.Context) {
	// Get current user
	user, err := p.userRepo.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user")
		log.Log.Errorf("failed to get current user: %v", err)
		return
	}

	rsp, err := p.repo.ListPreAuthKeyWithString(user.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(c, nil, "unknown error")
		log.Log.Errorf("list preauth key error: %v", err)
		return
	}
	//if err != nil && err.Error() != "rpc error: code = Unknown desc = User not found" {
	//	response.Fail(c, nil, "Not found user")
	//	log.Log.Errorf("not found user: %v", err)
	//	return
	//}
	response.Success(c, rsp, "Success")
}

// CreatePreAuthKey create PreAuthKey by current user
func (p *preAuthKeyController) CreatePreAuthKey(c *gin.Context) {
	var req vo.CreatePreAuthKey
	// Bind parameters
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Get current user
	user, err := p.userRepo.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "can't get user info")
		log.Log.Error(err)
		return
	}
	req.User = user.Name

	// Parse the ISO time format and convert it to timestamppb and reassign req.Expiration
	expire, err := time.Parse("2006-01-02T15:04:05.000Z", req.Expire)
	if err != nil {
		response.Fail(c, nil, "expire time format error")
		log.Log.Error(err)
		return
	}
	req.Expiration = timestamppb.New(expire)

	key, err := p.repo.CreatePreAuthKey(&req)
	if err != nil {
		response.Fail(c, nil, "Failed to create PreAuthKey")
		log.Log.Errorf("create PreAuthKey error: %v", err)
		return
	}
	response.Success(c, key, "success")
}

// ExpirePreAuthKey
func (p *preAuthKeyController) ExpirePreAuthKey(c *gin.Context) {
	var req vo.ExpirePreAuthKey
	// Bind parameters
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}

	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Get current user and reassign req.User
	user, err := p.userRepo.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "can't get user info")
		log.Log.Error(err)
		return
	}

	req.User = user.Name
	if err := p.repo.ExpirePreAuthKey(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	response.Success(c, nil, "success")
}
