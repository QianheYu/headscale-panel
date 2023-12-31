package repository

import (
	"errors"
	"fmt"
	"headscale-panel/common"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/vo"
	"strings"
)

type IRoleRepository interface {
	GetRoles(req *vo.RoleListRequest) ([]model.Role, int64, error)       // Get role list
	GetRolesByIds(roleIds []uint) ([]*model.Role, error)                 // Get role by role IDs
	CreateRole(role *model.Role) error                                   // Create role
	UpdateRoleById(roleId uint, role *model.Role) error                  // Update role
	GetRoleMenusById(roleId uint) ([]*model.Menu, error)                 // Get role's permission menu
	UpdateRoleMenus(role *model.Role) error                              // Update role's permission menu
	GetRoleApisByRoleKeyword(roleKeyword string) ([]*model.Api, error)   // Get role's permission API by role keyword
	UpdateRoleApis(roleKeyword string, reqRolePolicies [][]string) error // Update role's permission API (delete all first, then add)
	BatchDeleteRoleByIds(roleIds []uint) error                           // Delete role
}

type RoleRepository struct{}

func NewRoleRepository() IRoleRepository {
	return RoleRepository{}
}

// Get role list
func (r RoleRepository) GetRoles(req *vo.RoleListRequest) ([]model.Role, int64, error) {
	var list []model.Role
	db := common.DB.Model(&model.Role{}).Order("created_at DESC")

	name := strings.TrimSpace(req.Name)
	if name != "" {
		db = db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		db = db.Where("keyword LIKE ?", fmt.Sprintf("%%%s%%", keyword))
	}
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	// When pageNum > 0 and pageSize > 0, pagination is applied
	// Record total count
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return list, total, err
	}
	pageNum := int(req.PageNum)
	pageSize := int(req.PageSize)
	if pageNum > 0 && pageSize > 0 {
		err = db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error
	} else {
		err = db.Find(&list).Error
	}
	return list, total, err
}

// Get roles by role IDs
func (r RoleRepository) GetRolesByIds(roleIds []uint) ([]*model.Role, error) {
	var list []*model.Role
	err := common.DB.Where("id IN (?)", roleIds).Find(&list).Error
	return list, err
}

// Create role
func (r RoleRepository) CreateRole(role *model.Role) error {
	err := common.DB.Create(role).Error
	return err
}

// Update role
func (r RoleRepository) UpdateRoleById(roleId uint, role *model.Role) error {
	err := common.DB.Model(&model.Role{}).Where("id = ?", roleId).Updates(role).Error
	return err
}

// Get role's permission menu
func (r RoleRepository) GetRoleMenusById(roleId uint) ([]*model.Menu, error) {
	var role model.Role
	err := common.DB.Where("id = ?", roleId).Preload("Menus").First(&role).Error
	return role.Menus, err
}

func (r RoleRepository) GetRoleUsersById(roleId uint) ([]*model.User, error) {
	var role model.Role
	err := common.DB.Where("id = ?", roleId).Preload("Users").First(&role).Error
	return role.Users, err
}

// Update role's permission menu
func (r RoleRepository) UpdateRoleMenus(role *model.Role) error {
	err := common.DB.Model(role).Association("Menus").Replace(role.Menus)
	if err != nil {
		return err
	}
	go func() {
		if users, err := r.GetRoleUsersById(role.ID); err == nil {
			SetUsersRefreshFlag(users)
		} else {
			log.Log.Errorf("set role users error: %v", err)
		}
	}()
	return nil
}

// Get role's permission API by role keyword
func (r RoleRepository) GetRoleApisByRoleKeyword(roleKeyword string) ([]*model.Api, error) {
	policies := common.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)

	// Get all API
	var apis []*model.Api
	err := common.DB.Find(&apis).Error
	if err != nil {
		return apis, errors.New("get role's permission API failed")
	}

	accessApis := make([]*model.Api, 0)

	for _, policy := range policies {
		path := policy[1]
		method := policy[2]
		for _, api := range apis {
			if path == api.Path && method == api.Method {
				accessApis = append(accessApis, api)
				break
			}
		}
	}

	return accessApis, err

}

// Update role's permission API (delete all first, then add)
func (r RoleRepository) UpdateRoleApis(roleKeyword string, reqRolePolicies [][]string) error {
	// Get the existing police corresponding to the role ID in the path (to be deleted first)
	err := common.CasbinEnforcer.LoadPolicy()
	if err != nil {
		return errors.New("role's permission API strategy loading failed")
	}
	rmPolicies := common.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)
	if len(rmPolicies) > 0 {
		isRemoved, _ := common.CasbinEnforcer.RemovePolicies(rmPolicies)
		if !isRemoved {
			return errors.New("update role's permission API failed")
		}
	}
	isAdded, _ := common.CasbinEnforcer.AddPolicies(reqRolePolicies)
	if !isAdded {
		return errors.New("update role's permission API failed")
	}
	err = common.CasbinEnforcer.LoadPolicy()
	if err != nil {
		return errors.New("update role's permission API succeeded, role's permission API strategy loading failed")
	} else {
		return err
	}
}

// Delete role
func (r RoleRepository) BatchDeleteRoleByIds(roleIds []uint) error {
	var roles []*model.Role
	err := common.DB.Where("id IN (?)", roleIds).Find(&roles).Error
	if err != nil {
		return err
	}
	err = common.DB.Select("Users", "Menus").Unscoped().Delete(&roles).Error
	// Delete the casbin policy if successful
	if err == nil {
		for _, role := range roles {
			roleKeyword := role.Keyword
			rmPolicies := common.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)
			if len(rmPolicies) > 0 {
				isRemoved, _ := common.CasbinEnforcer.RemovePolicies(rmPolicies)
				if !isRemoved {
					return errors.New("delete role succeeded, delete role related permission API failed")
				}
			}
		}

	}
	return err
}
