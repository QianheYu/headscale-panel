package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/thoas/go-funk"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/vo"
	"strconv"
)

type IRoleController interface {
	GetRoles(c *gin.Context)             // Get role list
	CreateRole(c *gin.Context)           // Create role
	UpdateRoleById(c *gin.Context)       // Update role
	GetRoleMenusById(c *gin.Context)     // Get role menus by ID
	UpdateRoleMenusById(c *gin.Context)  // Update role menus by ID
	GetRoleApisById(c *gin.Context)      // Get role APIs by ID
	UpdateRoleApisById(c *gin.Context)   // Update role APIs by ID
	BatchDeleteRoleByIds(c *gin.Context) // Batch delete users
}

type RoleController struct {
	RoleRepository repository.IRoleRepository
}

func NewRoleController() IRoleController {
	roleRepository := repository.NewRoleRepository()
	roleController := RoleController{RoleRepository: roleRepository}
	return roleController
}

// Get role list
func (rc RoleController) GetRoles(c *gin.Context) {
	var req vo.RoleListRequest
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

	// Get roles
	roles, total, err := rc.RoleRepository.GetRoles(&req)
	if err != nil {
		response.Fail(c, nil, "Failed to get role list")
		log.Log.Errorf("get role list error: %v", err)
		return
	}
	response.Success(c, gin.H{"roles": roles, "total": total}, "Successfully got role list")
}

// Create role
func (rc RoleController) CreateRole(c *gin.Context) {
	var req vo.CreateRoleRequest
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

	// Get current user the highest role level
	uc := repository.NewUserRepository()
	sort, ctxUser, err := uc.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get the highest role level of the current user")
		log.Log.Errorf("get the highest role level of the current user error: %v", err)
		return
	}

	// User cannot create a role with a higher or equal level than self
	if sort >= req.Sort {
		response.Fail(c, nil, "Cannot create a role with a higher or equal level than yourself")
		return
	}

	role := model.Role{
		Name:    req.Name,
		Keyword: req.Keyword,
		Desc:    &req.Desc,
		Status:  req.Status,
		Sort:    req.Sort,
		Creator: ctxUser.Name,
	}

	// Create role
	err = rc.RoleRepository.CreateRole(&role)
	if err != nil {
		response.Fail(c, nil, "Failed to create role")
		log.Log.Errorf("create role error: %v", err)
		return
	}
	response.Success(c, nil, "Successfully created role")
}

// Update role
func (rc RoleController) UpdateRoleById(c *gin.Context) {
	var req vo.CreateRoleRequest
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
	// Get roleId from path
	roleId, _ := strconv.Atoi(c.Param("roleId"))
	if roleId <= 0 {
		response.Fail(c, nil, "Incorrect role ID")
		return
	}

	// Current user role sort minimum (highest ranked role) and current user
	ur := repository.NewUserRepository()
	minSort, ctxUser, err := ur.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get role")
		log.Log.Errorf("get current user role error: %v", err)
		return
	}

	// Can't update a character that is higher or equal to your character level
	// Get the role information based on the role ID in the path
	roles, err := rc.RoleRepository.GetRolesByIds([]uint{uint(roleId)})
	if err != nil {
		response.Fail(c, nil, "Failed to get role")
		log.Log.Errorf("get role by id error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Role information not obtained")
		return
	}
	if minSort >= roles[0].Sort {
		response.Fail(c, nil, "Cannot update a role with a higher or equal level than yourself")
		return
	}

	// Cannot update the character level to a higher level than the current user
	if minSort >= req.Sort {
		response.Fail(c, nil, "Cannot update the role level to be higher or equal to the current user's level")
		return
	}

	role := model.Role{
		Name:    req.Name,
		Keyword: req.Keyword,
		Desc:    &req.Desc,
		Status:  req.Status,
		Sort:    req.Sort,
		Creator: ctxUser.Name,
	}

	// Update role
	err = rc.RoleRepository.UpdateRoleById(uint(roleId), &role)
	if err != nil {
		response.Fail(c, nil, "Failed to update role")
		log.Log.Errorf("update role error: %v", err)
		return
	}

	// If the update is successful and the keyword of the role is updated,
	// then update the policy in the casbin
	if req.Keyword != roles[0].Keyword {
		// Get policy
		rolePolicies := common.CasbinEnforcer.GetFilteredPolicy(0, roles[0].Keyword)
		if len(rolePolicies) == 0 {
			response.Success(c, nil, "Successfully updated role")
			return
		}
		rolePoliciesCopy := make([][]string, 0)
		// Replace keyword
		for _, policy := range rolePolicies {
			policyCopy := make([]string, len(policy))
			copy(policyCopy, policy)
			rolePoliciesCopy = append(rolePoliciesCopy, policyCopy)
			policy[0] = req.Keyword
		}

		//gormadapter does not implement UpdatePolicies method, wait for gorm to update ---
		//isUpdated, _ := common.CasbinEnforcer.UpdatePolicies(rolePoliciesCopy, rolePolicies)
		//if !isUpdated {
		//	response.Fail(c, nil, "Successfully updated role, but failed to update the associated permission interface of the role keyword")
		//	return
		//}

		// Here you need to add and then delete (deleting and then adding will result in an error)
		isAdded, _ := common.CasbinEnforcer.AddPolicies(rolePolicies)
		if !isAdded {
			response.Fail(c, nil, "Successfully updated role, but failed to update the associated permission interface of the role keyword")
			return
		}
		isRemoved, _ := common.CasbinEnforcer.RemovePolicies(rolePoliciesCopy)
		if !isRemoved {
			response.Fail(c, nil, "Successfully updated role, but failed to update the associated permission interface of the role keyword")
			return
		}
		err := common.CasbinEnforcer.LoadPolicy()
		if err != nil {
			response.Fail(c, nil, "Successfully updated role, but failed to load the policy of the associated permission interface of the role keyword")
			return
		}

	}

	// There are two ways to successfully process the user information cache for a role: (the second method is used here because the number of users in a role can be large and the second method can spread the pressure on the database)
	// 1. You can help the user update the user information cache for the role they have, using the following method
	// err = ur.UpdateUserInfoCacheByRoleId(uint(roleId))
	// 2. Clear the cache directly and let the active users re-cache the latest user information themselves
	ur.ClearUserInfoCache()
	response.Success(c, nil, "Successfully updated role")
}

// Get the role's permission menu
func (rc RoleController) GetRoleMenusById(c *gin.Context) {
	// Get the roleId in the path
	roleId, _ := strconv.Atoi(c.Param("roleId"))
	if roleId <= 0 {
		response.Fail(c, nil, "Role ID is incorrect")
		return
	}
	menus, err := rc.RoleRepository.GetRoleMenusById(uint(roleId))
	if err != nil {
		response.Fail(c, nil, "Failed to get the role's permission menu")
		log.Log.Errorf("get the role's permission menu error: %v", err)
		return
	}
	response.Success(c, gin.H{"menus": menus}, "Successfully obtained the role's permission menu")
}

// Update the role's permission menu
func (rc RoleController) UpdateRoleMenusById(c *gin.Context) {
	var req vo.UpdateRoleMenusRequest
	// Bind parameter
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameter
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}
	// Get the role's permission menu
	roleId, _ := strconv.Atoi(c.Param("roleId"))
	if roleId <= 0 {
		response.Fail(c, nil, "Role ID is incorrect")
		return
	}
	// Get the role information based on the roleId in the path
	roles, err := rc.RoleRepository.GetRolesByIds([]uint{uint(roleId)})
	if err != nil {
		response.Fail(c, nil, "Failed to update role menu")
		log.Log.Errorf("get role by id error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Role information not obtained")
		return
	}

	// The current user's role has the smallest sort value (highest level role) and the current user
	ur := repository.NewUserRepository()
	minSort, ctxUser, err := ur.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to update role menu")
		log.Log.Errorf("get current user error: %v", err)
		return
	}

	// (Non-administrator) Cannot update the permission menu of roles with higher or equal level to your own role
	if minSort != 1 {
		if minSort >= roles[0].Sort {
			response.Fail(c, nil, "Cannot update the permission menu of roles with higher or equal level to your own role")
			log.Log.Errorf("Cannot update the permission menu of roles with higher or equal level to your own role")
			return
		}
	}

	// Get the permission menu owned by the current user
	mr := repository.NewMenuRepository()
	ctxUserMenus, err := mr.GetUserMenusByUserId(ctxUser.ID)
	if err != nil {
		response.Fail(c, nil, "Failed to get the accessible menu list of the current user")
		log.Log.Errorf("get the accessible menu list of the current user: %v", err)
		return
	}

	// Get the permission menu ID owned by the current user
	ctxUserMenusIds := make([]uint, 0)
	for _, menu := range ctxUserMenus {
		ctxUserMenusIds = append(ctxUserMenusIds, menu.ID)
	}

	// The front-end sends the latest MenuIds collection
	menuIds := req.MenuIds

	// The menu collection that the user needs to modify
	reqMenus := make([]*model.Menu, 0)

	// (Non-administrator) Cannot set the role's permission menu more than the permission menu owned by the current user
	if minSort != 1 {
		for _, id := range menuIds {
			if !funk.Contains(ctxUserMenusIds, id) {
				response.Fail(c, nil, "No permission to set the menu")
				return
			}
		}

		for _, id := range menuIds {
			for _, menu := range ctxUserMenus {
				if id == menu.ID {
					reqMenus = append(reqMenus, menu)
					break
				}
			}
		}
	} else {
		// Administrators set it arbitrarily
		// Query menu based on menuIds
		menus, err := mr.GetMenus()
		if err != nil {
			response.Fail(c, nil, "Failed to get the menu list")
			log.Log.Errorf("get the menu list error: %v", err)
			return
		}
		for _, menuId := range menuIds {
			for _, menu := range menus {
				if menuId == menu.ID {
					reqMenus = append(reqMenus, menu)
				}
			}
		}
	}

	roles[0].Menus = reqMenus

	err = rc.RoleRepository.UpdateRoleMenus(roles[0])
	if err != nil {
		response.Fail(c, nil, "Failed to update the role's permission menu")
		log.Log.Errorf("update the role's permission menu: %v", err)
		return
	}

	response.Success(c, nil, "Successfully updated the role's permission menu")
}

// Get the role's permission interface
func (rc RoleController) GetRoleApisById(c *gin.Context) {
	// Get roleId from path
	roleId, _ := strconv.Atoi(c.Param("roleId"))
	if roleId <= 0 {
		response.Fail(c, nil, "Role ID is incorrect")
		return
	}
	// Get the role information based on the roleId in the path
	roles, err := rc.RoleRepository.GetRolesByIds([]uint{uint(roleId)})
	if err != nil {
		response.Fail(c, nil, "Failed to get role's api")
		log.Log.Errorf("get roles by id error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Role information not obtained")
		return
	}
	// Get the policy in casbin based on the role keyword
	keyword := roles[0].Keyword
	apis, err := rc.RoleRepository.GetRoleApisByRoleKeyword(keyword)
	if err != nil {
		response.Fail(c, nil, "Failed to get role's api")
		log.Log.Errorf("get role api by id error: %v", err)
		return
	}
	response.Success(c, gin.H{"apis": apis}, "Successfully obtained the role's permission interface")
}

// Update role's permission API
func (rc RoleController) UpdateRoleApisById(c *gin.Context) {
	var req vo.UpdateRoleApisRequest
	// Bind parameter
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameter
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Get roleId in the path
	roleId, _ := strconv.Atoi(c.Param("roleId"))
	if roleId <= 0 {
		response.Fail(c, nil, "Role ID is incorrect")
		return
	}
	// Get the role information based on the role ID in the path
	roles, err := rc.RoleRepository.GetRolesByIds([]uint{uint(roleId)})
	if err != nil {
		response.Fail(c, nil, "Failed to update role apis")
		log.Log.Errorf("ger roles by id error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Role information not obtained")
		return
	}

	// The smallest role sorting value (the highest level role) and the current user
	ur := repository.NewUserRepository()
	minSort, ctxUser, err := ur.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to update role apis")
		log.Log.Errorf("get current user error: %v", err)
		return
	}

	// (Non-administrator) cannot update the permission API of roles with higher or equal level to your own role
	if minSort != 1 {
		if minSort >= roles[0].Sort {
			response.Fail(c, nil, "Cannot update the permission API of roles with higher or equal level to your own role")
			return
		}
	}

	// Get the permission API owned by the current user
	ctxRoles := ctxUser.Roles
	ctxRolesPolicies := make([][]string, 0)
	for _, role := range ctxRoles {
		policy := common.CasbinEnforcer.GetFilteredPolicy(0, role.Keyword)
		ctxRolesPolicies = append(ctxRolesPolicies, policy...)
	}
	// Get the collection of permission interfaces that the role corresponding to the role ID in the path can set
	for _, policy := range ctxRolesPolicies {
		policy[0] = roles[0].Keyword
	}

	// The front-end sends the latest Api ID collection
	apiIds := req.ApiIds
	// Get interface details based on api ID
	ar := repository.NewApiRepository()
	apis, err := ar.GetApisById(apiIds)
	if err != nil {
		response.Fail(c, nil, "Failed to get interface information by API ID")
		return
	}
	// Generate the role polices you want to set up on the front end
	reqRolePolicies := make([][]string, 0)
	for _, api := range apis {
		reqRolePolicies = append(reqRolePolicies, []string{
			roles[0].Keyword, api.Path, api.Method,
		})
	}

	// (Non-administrator) cannot set the role's permission interface more than the permission interface owned by the current user
	if minSort != 1 {
		for _, reqPolicy := range reqRolePolicies {
			if !funk.Contains(ctxRolesPolicies, reqPolicy) {
				response.Fail(c, nil, fmt.Sprintf("No permission to set the interface with path %s and request method %s", reqPolicy[1], reqPolicy[2]))
				return
			}
		}
	}

	// Permissions API for updating roles
	err = rc.RoleRepository.UpdateRoleApis(roles[0].Keyword, reqRolePolicies)
	if err != nil {
		response.Fail(c, nil, "Failed to update role apis")
		log.Log.Errorf("update role apis error: %v", err)
		return
	}

	response.Success(c, nil, "Successfully updated the role's permission interface")
}

// Batch delete roles
func (rc RoleController) BatchDeleteRoleByIds(c *gin.Context) {
	var req vo.DeleteRoleRequest
	// Bind parameter
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, nil, "param error")
		return
	}
	// Validate parameter
	if err := common.Validate.Struct(&req); err != nil {
		errStr := err.(validator.ValidationErrors)[0].Translate(common.Trans)
		response.Fail(c, nil, errStr)
		return
	}

	// Get the highest level role of the current user
	ur := repository.NewUserRepository()
	minSort, _, err := ur.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to delete role")
		log.Log.Errorf("get current user error: %v", err)
		return
	}

	// The front-end sends the role IDs to be deleted
	roleIds := req.RoleIds
	// Get role information
	roles, err := rc.RoleRepository.GetRolesByIds(roleIds)
	if err != nil {
		response.Fail(c, nil, "Failed to delete role")
		log.Log.Errorf("get role information error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Cannot delete roles with higher or equal level to your own role")
		return
	}

	// Cannot delete roles with higher or equal level to your own role
	for _, role := range roles {
		if minSort >= role.Sort {
			response.Fail(c, nil, "Cannot delete roles with higher or equal level to your own role")
			return
		}
	}

	// Delete role
	err = rc.RoleRepository.BatchDeleteRoleByIds(roleIds)
	if err != nil {
		response.Fail(c, nil, "Failed to delete role")
		return
	}

	// After successfully deleting the role, clear the cache directly and let active users re-cache the latest user information themselves
	ur.ClearUserInfoCache()
	response.Success(c, nil, "Successfully deleted role")
}
