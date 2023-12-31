package repository

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"headscale-panel/common"
	"headscale-panel/model"
	"headscale-panel/util"
	"headscale-panel/vo"
	"strings"
	"time"
)

type IUserRepository interface {
	Login(user *model.User) (*model.User, error)   // Login
	ChangePwd(name string, newPasswd string) error // Update password

	CreateUser(user *model.User) error                              // Create user
	GetUserById(id uint) (model.User, error)                        // Get single user
	GetUsers(req *vo.UserListRequest) ([]*model.User, int64, error) // Get user list
	UpdateUser(user *model.User) error                              // Update user
	BatchDeleteUserByIds(ids []uint) error                          // Batch delete users
	BatchDeleteUserByNames(name []string) error

	GetCurrentUser(c *gin.Context) (model.User, error)                  // Get current user information
	GetCurrentUserMinRoleSort(c *gin.Context) (uint, model.User, error) // Get the minimum role sorting value (highest level role) and current user information of the current user
	GetUserMinRoleSortsByIds(ids []uint) ([]int, error)                 // Get the minimum role sorting value by user ID

	SetUserInfoCache(name string, user model.User) // Set user information cache
	UpdateUserInfoCacheByRoleId(roleId uint) error // Update the user information cache for users with the role ID
	ClearUserInfoCache()                           // Clear all user information caches
}

type UserRepository struct{}

// Cache current user information to avoid frequent database access
var userInfoCache = cache.New(24*time.Hour, 48*time.Hour)

func SetRefreshToken() {
	for key, item := range userInfoCache.Items() {
		user := item.Object.(model.User)
		user.RefreshFlag = false
		userInfoCache.Set(key, user, cache.DefaultExpiration)
	}
}

func SetUsersRefreshFlag(users []*model.User) {
	for _, user := range users {
		SetUserRefreshFlag(user)
	}
}

func SetUserRefreshFlag(user *model.User) {
	user.RefreshFlag = true
	userInfoCache.Set(user.Name, *user, cache.DefaultExpiration)
}

func ClearUserRefreshToken(username string) {
	if v, ok := userInfoCache.Get(username); ok {
		user := v.(model.User)
		user.RefreshFlag = false
		userInfoCache.Set(username, user, cache.DefaultExpiration)
	}
}

func GetUserRefreshToken(username string) bool {
	if v, ok := userInfoCache.Get(username); ok && v.(model.User).RefreshFlag {
		return true
	}
	return false
}

func NewUserRepository() IUserRepository {
	return UserRepository{}
}

// Login
func (ur UserRepository) Login(user *model.User) (*model.User, error) {
	// Get user by username (normal status: user status is normal)
	var firstUser model.User
	err := common.DB.
		Where("name = ?", user.Name).
		Preload("Roles").
		First(&firstUser).Error
	if err != nil {
		return nil, errors.New("user does not exist")
	}

	// Determine the user's status
	userStatus := firstUser.Status
	if userStatus != 1 {
		return nil, errors.New("user is disabled")
	}

	// Determine the status of all roles owned by the user, and if all roles are disabled, the user cannot log in
	roles := firstUser.Roles
	isValidate := false
	for _, role := range roles {
		// If there is a role with a normal status, the user can log in
		if role.Status == 1 {
			isValidate = true
			break
		}
	}

	if !isValidate {
		return nil, errors.New("user role is disabled")
	}

	// Verify password
	err = util.ComparePasswd(firstUser.Password, user.Password)
	if err != nil {
		return &firstUser, errors.New("wrong password")
	}
	//userInfoCache.Set(firstUser.Name, firstUser, cache.DefaultExpiration)
	return &firstUser, nil
}

// Get current logged-in user information
// Need to cache to reduce database access
func (ur UserRepository) GetCurrentUser(c *gin.Context) (model.User, error) {
	var newUser model.User
	ctxUser, exist := c.Get("user")
	if !exist {
		return newUser, errors.New("user not logged in")
	}
	u, _ := ctxUser.(model.User)

	// First, get the cache
	cacheUser, found := userInfoCache.Get(u.Name)
	var user model.User
	var err error
	if found {
		user = cacheUser.(model.User)
		err = nil
	} else {
		// If there is no cache, get the data from the database
		user, err = ur.GetUserById(u.ID)
		// Cache if the data is retrieved successfully
		if err != nil {
			userInfoCache.Delete(u.Name)
		} else {
			userInfoCache.Set(u.Name, user, cache.DefaultExpiration)
		}
	}
	return user, err
}

// Get the minimum role sorting value (highest level role) and current user information of the current user
func (ur UserRepository) GetCurrentUserMinRoleSort(c *gin.Context) (uint, model.User, error) {
	// Get current user
	ctxUser, err := ur.GetCurrentUser(c)
	if err != nil {
		return 999, ctxUser, err
	}
	// get all roles of the current user
	currentRoles := ctxUser.Roles
	// Get the sorting of the current user's roles and compare it with the sorting of the roles sent by the frontend
	var currentRoleSorts []int
	for _, role := range currentRoles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}
	// Minimum sorting value of the current user's roles (highest level role)
	//currentRoleSortMin := uint(funk.MinInt(currentRoleSorts).(int))
	currentRoleSortMin := uint(funk.MinInt(currentRoleSorts))
	return currentRoleSortMin, ctxUser, nil
}

// Get single user
func (ur UserRepository) GetUserById(id uint) (model.User, error) {
	var user model.User
	err := common.DB.Where("id = ?", id).Preload("Roles").First(&user).Error
	return user, err
}

// Get user list
func (ur UserRepository) GetUsers(req *vo.UserListRequest) ([]*model.User, int64, error) {
	var list []*model.User
	db := common.DB.Model(&model.User{}).Order("created_at DESC")

	username := strings.TrimSpace(req.Username)
	if username != "" {
		db = db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", username))
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname != "" {
		db = db.Where("nickname LIKE ?", fmt.Sprintf("%%%s%%", nickname))
	}
	//mobile := strings.TrimSpace(req.Mobile)
	//if mobile != "" {
	//	db = db.Where("mobile LIKE ?", fmt.Sprintf("%%%s%%", mobile))
	//}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		db = db.Where("email LIKE ?", fmt.Sprintf("%%%s%%", email))
	}
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	// Pagination occurs only when pageNum > 0 and pageSize > 0
	// Record the total number of items
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return list, total, err
	}
	pageNum := int(req.PageNum)
	pageSize := int(req.PageSize)
	if pageNum > 0 && pageSize > 0 {
		err = db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Preload("Roles").Find(&list).Error
	} else {
		err = db.Preload("Roles").Find(&list).Error
	}
	return list, total, err
}

// Update password
func (ur UserRepository) ChangePwd(name string, hashNewPasswd string) error {
	err := common.DB.Model(&model.User{}).Where("name = ?", name).Update("password", hashNewPasswd).Error
	// If the password is updated successfully, update the current user information cache
	// First, get the cache
	cacheUser, found := userInfoCache.Get(name)
	if err == nil {
		if found {
			user := cacheUser.(model.User)
			user.Password = hashNewPasswd
			userInfoCache.Set(name, user, cache.DefaultExpiration)
		} else {
			// If there is no cache, get the user information cache
			var user model.User
			common.DB.Where("name = ?", name).First(&user)
			userInfoCache.Set(name, user, cache.DefaultExpiration)
		}
	}

	return err
}

// Create user
func (ur UserRepository) CreateUser(user *model.User) error {
	err := common.DB.Create(user).Error
	return err
}

// Update user
func (ur UserRepository) UpdateUser(user *model.User) error {
	err := common.DB.Model(user).Save(user).Error
	if err != nil {
		return err
	}
	err = common.DB.Model(user).Association("Roles").Replace(user.Roles)

	//err := common.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&user).Error

	// Update the user information cache if the update is successful
	if err == nil {
		SetUserRefreshFlag(user)
	}
	return err
}

// Batch delete
func (ur UserRepository) BatchDeleteUserByIds(ids []uint) error {
	// Users and roles have a many-to-many relationship
	var users []model.User
	for _, id := range ids {
		// Get user by ID
		user, err := ur.GetUserById(id)
		if err != nil && err != gorm.ErrRecordNotFound {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return fmt.Errorf(" Get user with ID %d not found", id)
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return nil
	}
	err := common.DB.Select("Roles").Unscoped().Delete(&users).Error
	// If the user is deleted successfully, delete the user information cache
	if err == nil {
		for _, user := range users {
			userInfoCache.Delete(user.Name)
		}
	}
	return err
}

func (ur UserRepository) BatchDeleteUserByNames(name []string) error {
	user := &model.User{}
	return common.DB.Model(user).Where("user_name in (?)", name).Delete(user).Error
}

// Get the minimum role sorting value by user ID
func (ur UserRepository) GetUserMinRoleSortsByIds(ids []uint) ([]int, error) {
	// Get user information based on user ID
	var userList []model.User
	err := common.DB.Where("id IN (?)", ids).Preload("Roles").Find(&userList).Error
	if err != nil {
		return []int{}, err
	}
	if len(userList) == 0 {
		return []int{}, errors.New("no user information found")
	}
	var roleMinSortList []int
	for _, user := range userList {
		roles := user.Roles
		var roleSortList []int
		for _, role := range roles {
			roleSortList = append(roleSortList, int(role.Sort))
		}
		if len(roleSortList) == 0 {
			// The user has no role information, directly continue
			continue
		}
		roleMinSort := funk.MinInt(roleSortList)
		roleMinSortList = append(roleMinSortList, roleMinSort)
	}
	return roleMinSortList, nil
}

// Set user information cache
func (ur UserRepository) SetUserInfoCache(name string, user model.User) {
	userInfoCache.Set(name, user, cache.DefaultExpiration)
}

// Update the user information cache for users with the role ID
func (ur UserRepository) UpdateUserInfoCacheByRoleId(roleId uint) error {

	var role model.Role
	err := common.DB.Where("id = ?", roleId).Preload("Users").First(&role).Error
	if err != nil {
		return errors.New("failed to get role information by role ID")
	}

	users := role.Users
	if len(users) == 0 {
		return errors.New("the user with the role was not retrieved based on the role ID")
	}

	// Update user information cache
	for _, user := range users {
		_, found := userInfoCache.Get(user.Name)
		if found {
			userInfoCache.Set(user.Name, *user, cache.DefaultExpiration)
		}
	}

	return err
}

// Clear all user information caches
func (ur UserRepository) ClearUserInfoCache() {
	userInfoCache.Flush()
}
