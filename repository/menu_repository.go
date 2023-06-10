package repository

import (
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"headscale-panel/common"
	"headscale-panel/model"
)

type IMenuRepository interface {
	GetMenus() ([]*model.Menu, error)                   // Get menu list
	GetMenuTree() ([]*model.Menu, error)                // Get menu tree
	CreateMenu(menu *model.Menu) error                  // Create menu
	UpdateMenuById(menuId uint, menu *model.Menu) error // Update menu
	BatchDeleteMenuByIds(menuIds []uint) error          // Batch delete menus

	GetUserMenusByUserId(userId uint) ([]*model.Menu, error)    // Get user's access menu list by user ID
	GetUserMenuTreeByUserId(userId uint) ([]*model.Menu, error) // Get user's access menu tree by user ID
}

type MenuRepository struct{}

func NewMenuRepository() IMenuRepository {
	return MenuRepository{}
}

// Get menu list
func (m MenuRepository) GetMenus() ([]*model.Menu, error) {
	var menus []*model.Menu
	err := common.DB.Order("sort").Find(&menus).Error
	return menus, err
}

// Get menu tree
func (m MenuRepository) GetMenuTree() ([]*model.Menu, error) {
	var menus []*model.Menu
	err := common.DB.Order("sort").Find(&menus).Error
	// parentId为0的是根菜单
	return GenMenuTree(0, menus), err
}

func GenMenuTree(parentId uint, menus []*model.Menu) []*model.Menu {
	tree := make([]*model.Menu, 0)

	for _, m := range menus {
		if m.ParentId == parentId {
			m.Children = GenMenuTree(m.ID, menus)
			tree = append(tree, m)
		}
	}
	return tree
}

// Create menu
func (m MenuRepository) CreateMenu(menu *model.Menu) error {
	err := common.DB.Create(menu).Error
	return err
}

// Update menu
func (m MenuRepository) UpdateMenuById(menuId uint, menu *model.Menu) error {
	err := common.DB.Model(menu).Where("id = ?", menuId).Updates(menu).Error
	return err
}

// Batch delete menus
func (m MenuRepository) BatchDeleteMenuByIds(menuIds []uint) error {
	var menus []*model.Menu
	err := common.DB.Where("id IN (?)", menuIds).Find(&menus).Error
	if err != nil {
		return err
	}
	err = common.DB.Select("Roles").Unscoped().Delete(&menus).Error
	return err
}

// Get user's access menu list by user ID
func (m MenuRepository) GetUserMenusByUserId(userId uint) ([]*model.Menu, error) {
	// Get user
	var user model.User
	err := common.DB.Where("id = ?", userId).Preload("Roles").First(&user).Error
	if err != nil {
		return nil, err
	}
	// Get roles
	roles := user.Roles
	// All roles' menu collection
	allRoleMenus := make([]*model.Menu, 0)
	for _, role := range roles {
		var userRole model.Role
		// I hope to sort by the sort field in the Menus table from small to large
		err := common.DB.Where("id = ?", role.ID).Preload("Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}).First(&userRole).Error
		if err != nil {
			return nil, err
		}
		// Get role's menu
		menus := userRole.Menus
		allRoleMenus = append(allRoleMenus, menus...)
	}

	// Remove duplicates from all roles' menu collection
	allRoleMenusId := make([]int, 0)
	for _, menu := range allRoleMenus {
		allRoleMenusId = append(allRoleMenusId, int(menu.ID))
	}
	allRoleMenusIdUniq := funk.UniqInt(allRoleMenusId)
	allRoleMenusUniq := make([]*model.Menu, 0)
	for _, id := range allRoleMenusIdUniq {
		for _, menu := range allRoleMenus {
			if id == int(menu.ID) {
				allRoleMenusUniq = append(allRoleMenusUniq, menu)
				break
			}
		}
	}

	// Get menus with status 1
	accessMenus := make([]*model.Menu, 0)
	for _, menu := range allRoleMenusUniq {
		if menu.Status == 1 {
			accessMenus = append(accessMenus, menu)
		}
	}

	return accessMenus, err
}

// Get user's access menu tree by user ID
func (m MenuRepository) GetUserMenuTreeByUserId(userId uint) ([]*model.Menu, error) {
	menus, err := m.GetUserMenusByUserId(userId)
	if err != nil {
		return nil, err
	}
	tree := GenMenuTree(0, menus)
	return tree, err
}
