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
	"strconv"
)

type IApiController interface {
	GetApis(c *gin.Context)             // Get Api list
	GetApiTree(c *gin.Context)          // Get Api tree (classified ty interface Category field)
	CreateApi(c *gin.Context)           // Create Api
	UpdateApiById(c *gin.Context)       // Update Api
	BatchDeleteApiByIds(c *gin.Context) // Batch delete Apis
}

type ApiController struct {
	ApiRepository repository.IApiRepository
}

func NewApiController() IApiController {
	apiRepository := repository.NewApiRepository()
	apiController := ApiController{ApiRepository: apiRepository}
	return apiController
}

// Get Api list
func (ac ApiController) GetApis(c *gin.Context) {
	var req vo.ApiListRequest
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}
	// Get Apis
	apis, total, err := ac.ApiRepository.GetApis(&req)
	if err != nil {
		response.Fail(c, nil, "Failed to get API list")
		return
	}
	response.Success(c, gin.H{
		"apis": apis, "total": total,
	}, "Get API list successfully")
}

// Get Api tree (classified by Api Category field)
func (ac ApiController) GetApiTree(c *gin.Context) {
	tree, err := ac.ApiRepository.GetApiTree()
	if err != nil {
		response.Fail(c, nil, "Failed to get API tree")
		return
	}
	response.Success(c, gin.H{
		"apiTree": tree,
	}, "Get API tree successfully")
}

// Create Api
func (ac ApiController) CreateApi(c *gin.Context) {
	var req vo.CreateApiRequest
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
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
	ur := repository.NewUserRepository()
	ctxUser, err := ur.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user information")
		return
	}

	api := model.Api{
		Method:   req.Method,
		Path:     req.Path,
		Category: req.Category,
		Desc:     req.Desc,
		Creator:  ctxUser.Name,
	}

	// Create API
	err = ac.ApiRepository.CreateApi(&api)
	if err != nil {
		response.Fail(c, nil, "Failed to create API")
		log.Log.Errorf("create api error: %v", err)
		return
	}

	response.Success(c, nil, "Create API successfully")
}

// Update Api
func (ac ApiController) UpdateApiById(c *gin.Context) {
	var req vo.UpdateApiRequest
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Get apiId in path
	apiId, _ := strconv.Atoi(c.Param("apiId"))
	if apiId <= 0 {
		response.Fail(c, nil, "Incorrect API ID")
		return
	}

	// Get current user
	ur := repository.NewUserRepository()
	ctxUser, err := ur.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user information")
		return
	}

	api := model.Api{
		Method:   req.Method,
		Path:     req.Path,
		Category: req.Category,
		Desc:     req.Desc,
		Creator:  ctxUser.Name,
	}

	err = ac.ApiRepository.UpdateApiById(uint(apiId), &api)
	if err != nil {
		response.Fail(c, nil, "Failed to update API")
		log.Log.Errorf("update api error: %v", err)
		return
	}

	response.Success(c, nil, "Update API successfully")
}

// Batch delete Apis
func (ac ApiController) BatchDeleteApiByIds(c *gin.Context) {
	var req vo.DeleteApiRequest
	// Bind parameters
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameters
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Delete Apis
	err := ac.ApiRepository.BatchDeleteApiByIds(req.ApiIds)
	if err != nil {
		response.Fail(c, nil, "Failed to delete API")
		log.Log.Errorf("delete api error: %v", err)
		return
	}

	response.Success(c, nil, "Delete API successfully")
}
