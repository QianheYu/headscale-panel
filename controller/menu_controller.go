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

type IMenuController interface {
	GetMenus(c *gin.Context)             // Get menu list
	GetMenuTree(c *gin.Context)          // Get menu tree
	CreateMenu(c *gin.Context)           // Create menu
	UpdateMenuById(c *gin.Context)       // Update menu
	BatchDeleteMenuByIds(c *gin.Context) // Batch delete menus

	GetUserMenusByUserId(c *gin.Context)    // Get user's accessible menu list
	GetUserMenuTreeByUserId(c *gin.Context) // Get user's accessible menu tree
}

type MenuController struct {
	MenuRepository repository.IMenuRepository
	UserRepository repository.IUserRepository
}

func NewMenuController() IMenuController {
	menuController := MenuController{MenuRepository: repository.NewMenuRepository(), UserRepository: repository.NewUserRepository()}
	return menuController
}

// Get menu list
func (mc MenuController) GetMenus(c *gin.Context) {
	menus, err := mc.MenuRepository.GetMenus()
	if err != nil {
		response.Fail(c, nil, "Failed to get menu list")
		log.Log.Errorf("get menu list error: %v", err)
		return
	}
	response.Success(c, gin.H{"menus": menus}, "Successfully got menu list")
}

// Get menu tree
func (mc MenuController) GetMenuTree(c *gin.Context) {
	menuTree, err := mc.MenuRepository.GetMenuTree()
	if err != nil {
		response.Fail(c, nil, "Failed to get menu tree")
		log.Log.Errorf("get menu tree error: %v", err)
		return
	}
	response.Success(c, gin.H{"menuTree": menuTree}, "Successfully got menu tree")
}

// Create menu
func (mc MenuController) CreateMenu(c *gin.Context) {
	var req vo.CreateMenuRequest
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
	ctxUser, err := mc.UserRepository.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user information")
		return
	}

	menu := model.Menu{
		Name:       req.Name,
		Title:      req.Title,
		Icon:       &req.Icon,
		Path:       req.Path,
		Redirect:   &req.Redirect,
		Component:  req.Component,
		Sort:       req.Sort,
		Status:     req.Status,
		Hidden:     req.Hidden,
		Cache:      req.Cache,
		AlwaysShow: req.AlwaysShow,
		Breadcrumb: req.Breadcrumb,
		ActiveMenu: &req.ActiveMenu,
		ParentId:   req.ParentId,
		Creator:    ctxUser.Name,
	}

	err = mc.MenuRepository.CreateMenu(&menu)
	if err != nil {
		response.Fail(c, nil, "Create menu failed")
		log.Log.Errorf("create menu error: %v", err)
		return
	}
	response.Success(c, nil, "Successfully created menu")
}

// Update menu
func (mc MenuController) UpdateMenuById(c *gin.Context) {
	var req vo.UpdateMenuRequest
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

	// Get menuId in path
	menuId, _ := strconv.Atoi(c.Param("menuId"))
	if menuId <= 0 {
		response.Fail(c, nil, "Incorrect menu ID")
		return
	}

	// Get current user
	ctxUser, err := mc.UserRepository.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user information")
		return
	}

	menu := model.Menu{
		Name:       req.Name,
		Title:      req.Title,
		Icon:       &req.Icon,
		Path:       req.Path,
		Redirect:   &req.Redirect,
		Component:  req.Component,
		Sort:       req.Sort,
		Status:     req.Status,
		Hidden:     req.Hidden,
		Cache:      req.Cache,
		AlwaysShow: req.AlwaysShow,
		Breadcrumb: req.Breadcrumb,
		ActiveMenu: &req.ActiveMenu,
		ParentId:   req.ParentId,
		Creator:    ctxUser.Name,
	}

	err = mc.MenuRepository.UpdateMenuById(uint(menuId), &menu)
	if err != nil {
		response.Fail(c, nil, "Update menu failed")
		log.Log.Errorf("update menu error: %v", err)
		return
	}

	response.Success(c, nil, "Successfully updated menu")

}

// Batch delete menus
func (mc MenuController) BatchDeleteMenuByIds(c *gin.Context) {
	var req vo.DeleteMenuRequest
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
	err := mc.MenuRepository.BatchDeleteMenuByIds(req.MenuIds)
	if err != nil {
		response.Fail(c, nil, "Failed to delete menu")
		log.Log.Errorf("delete menu error: %v", err)
		return
	}

	response.Success(c, nil, "Successfully deleted menu")
}

// Get a list of the user's accessible menus list based on their user ID
func (mc MenuController) GetUserMenusByUserId(c *gin.Context) {
	// Get userId in path
	userId, _ := strconv.Atoi(c.Param("userId"))
	if userId <= 0 {
		response.Fail(c, nil, "Incorrect user ID")
		return
	}

	menus, err := mc.MenuRepository.GetUserMenusByUserId(uint(userId))
	if err != nil {
		response.Fail(c, nil, "Failed to get user's accessible menu list")
		log.Log.Errorf("get user's accessible menu list error: %v", err)
		return
	}
	response.Success(c, gin.H{"menus": menus}, "Successfully got user's accessible menu list")
}

// Get the user's accessible menu tree based on the user ID
func (mc MenuController) GetUserMenuTreeByUserId(c *gin.Context) {
	// Get userId in path
	userId, _ := strconv.Atoi(c.Param("userId"))
	if userId <= 0 {
		response.Fail(c, nil, "Incurrect user ID")
		return
	}

	menuTree, err := mc.MenuRepository.GetUserMenuTreeByUserId(uint(userId))
	if err != nil {
		response.Fail(c, nil, "Failed to get user's accessible menu tree")
		log.Log.Errorf("get user's accessible menu tree error: %v", err)
		return
	}
	response.Success(c, gin.H{"menuTree": menuTree}, "Successfully got user's accessible menu tree")
}
