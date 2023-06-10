package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/thoas/go-funk"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/dto"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/repository"
	"headscale-panel/response"
	"headscale-panel/util"
	"headscale-panel/vo"
	"strconv"
)

type IUserController interface {
	GetUserInfo(c *gin.Context)          // Get currnent logged-in user information
	GetUsers(c *gin.Context)             // Get user list
	ChangePwd(c *gin.Context)            // Update user login password
	CreateUser(c *gin.Context)           // Create user
	UpdateUserById(c *gin.Context)       // Update user
	BatchDeleteUserByIds(c *gin.Context) // Batch delete users
}

type UserController struct {
	UserRepository          repository.IUserRepository
	HeadscaleUserRepository repository.HeadscaleUserRepository
}

func NewUserController() IUserController {
	userRepository := repository.NewUserRepository()
	headscaleUserRepository := repository.NewUserRepo()
	userController := UserController{UserRepository: userRepository, HeadscaleUserRepository: headscaleUserRepository}
	return userController
}

// Get current logged-in user information
func (uc UserController) GetUserInfo(c *gin.Context) {
	user, err := uc.UserRepository.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user information")
		log.Log.Errorf("get current user information error: %v", err)
		return
	}
	userInfoDto := dto.ToUserInfoDto(user)
	response.Success(c, gin.H{
		"userInfo": userInfoDto,
	}, "Successfully got current user information")
}

// Get user list
func (uc UserController) GetUsers(c *gin.Context) {
	var req vo.UserListRequest
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

	// Get user list
	users, total, err := uc.UserRepository.GetUsers(&req)
	if err != nil {
		response.Fail(c, nil, "Failed to get user list")
		log.Log.Errorf("get user list error: %v", err)
		return
	}
	response.Success(c, gin.H{"users": dto.ToUsersDto(users), "total": total}, "Successfully got user list")
}

// Update user login password
func (uc UserController) ChangePwd(c *gin.Context) {
	var req vo.ChangePwdRequest

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

	// The password from the front end is encrypted by rsa, decrypted first
	// The password is decrypted by RSA
	decodeOldPassword, err := util.RSADecrypt([]byte(req.OldPassword), config.Conf.System.PrivateKey)
	if err != nil {
		response.Fail(c, nil, "Operate password error")
		log.Log.Error(err)
		return
	}
	decodeNewPassword, err := util.RSADecrypt([]byte(req.NewPassword), config.Conf.System.PrivateKey)
	if err != nil {
		response.Fail(c, nil, "Operate password error")
		log.Log.Error(err)
		return
	}
	req.OldPassword = string(decodeOldPassword)
	req.NewPassword = string(decodeNewPassword)

	// Get current user
	user, err := uc.UserRepository.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Get current user failed")
		return
	}
	// Obtain the user's true and correct password
	correctPasswd := user.Password
	// Determine if the password requested by the front-end is equal to the real password
	err = util.ComparePasswd(correctPasswd, req.OldPassword)
	if err != nil {
		response.Fail(c, nil, "The original password is incorrect")
		return
	}
	// Update password
	err = uc.UserRepository.ChangePwd(user.Name, util.GenPasswd(req.NewPassword))
	if err != nil {
		response.Fail(c, nil, "Failed to update password")
		log.Log.Errorf("update password error: %v", err)
		return
	}
	response.Success(c, nil, "Successfully updated password")
}

// Create user
func (uc UserController) CreateUser(c *gin.Context) {
	var req vo.CreateUserRequest
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

	// Password decrypted by RSA
	// decryption if password is not empty
	if req.Password != "" {
		decodeData, err := util.RSADecrypt([]byte(req.Password), config.Conf.System.PrivateKey)
		if err != nil {
			response.Fail(c, nil, "Operate password error")
			log.Log.Error(err)
			return
		}
		req.Password = string(decodeData)
		if len(req.Password) < 6 {
			response.Fail(c, nil, "Password length must be at least 6 characters")
			return
		}
	}

	if req.Avatar == "" {
		req.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
	}

	// Current user role sort minimum (highest ranked role) and current user
	currentRoleSortMin, ctxUser, err := uc.UserRepository.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to current user role")
		log.Log.Error(err)
		return
	}

	// Get the user role id from the front end
	reqRoleIds := req.RoleIds
	// Get role based on role id
	rr := repository.NewRoleRepository()
	roles, err := rr.GetRolesByIds(reqRoleIds)
	if err != nil {
		response.Fail(c, nil, "Failed to get role information")
		log.Log.Errorf("get role information by role ID error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Failed to get role information")
		return
	}
	var reqRoleSorts []int
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}
	// Front end passes in user role sorting min (highest ranked role)
	reqRoleSortMin := uint(funk.MinInt(reqRoleSorts))

	// The current user's role sort minimum needs to be less than
	//the role sort minimum passed from the front end
	//(users cannot create users of a higher rank than their own or of the same rank)
	if currentRoleSortMin >= reqRoleSortMin {
		response.Fail(c, nil, "Users cannot create users with higher or equal levels than themselves")
		return
	}

	// Default 123456 if password is empty
	if req.Password == "" {
		req.Password = "123456"
	}
	user := model.User{
		Name:     req.Username,
		Password: util.GenPasswd(req.Password),
		//Mobile:       req.Mobile,
		Email:        req.Email,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       req.Status,
		Creator:      ctxUser.Name,
		Roles:        roles,
	}

	err = uc.UserRepository.CreateUser(&user)
	if err != nil {
		response.Fail(c, nil, "Failed to create user")
		log.Log.Errorf("create user error: %v", err)
		return
	}
	response.Success(c, nil, "Successfully created user")
}

// Update user
func (uc UserController) UpdateUserById(c *gin.Context) {
	var req vo.CreateUserRequest
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

	// Get the userId in the path
	userId, _ := strconv.Atoi(c.Param("userId"))
	if userId <= 0 {
		response.Fail(c, nil, "Incorrect user ID")
		return
	}

	// Get the user information based on the userId in the path
	oldUser, err := uc.UserRepository.GetUserById(uint(userId))
	if err != nil {
		response.Fail(c, nil, "Failed to get user information")
		log.Log.Errorf("get user information to update error: %v", err)
		return
	}

	// Get current user
	ctxUser, err := uc.UserRepository.GetCurrentUser(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user")
		log.Log.Error(err)
		return
	}
	// Get all roles of the current user
	currentRoles := ctxUser.Roles
	// Get the sorting of the current user roles and compare it
	// with the sorting of the roles from the front-end
	var currentRoleSorts []int
	// Collection of current user role IDs
	var currentRoleIds []uint
	for _, role := range currentRoles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
		currentRoleIds = append(currentRoleIds, role.ID)
	}
	// Current user role sort minimum (highest ranked role)
	//currentRoleSortMin := funk.MinInt(currentRoleSorts).(int)
	currentRoleSortMin := funk.MinInt(currentRoleSorts)

	// Get the user role id from the front end
	reqRoleIds := req.RoleIds
	// Get role based on role id
	rr := repository.NewRoleRepository()
	roles, err := rr.GetRolesByIds(reqRoleIds)
	if err != nil {
		response.Fail(c, nil, "Failed to get role")
		log.Log.Errorf("get role information by role ID error: %v", err)
		return
	}
	if len(roles) == 0 {
		response.Fail(c, nil, "Failed to get role information")
		return
	}
	var reqRoleSorts []int
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}
	// Front end passes in user role sorting min (highest ranked role)
	reqRoleSortMin := funk.MinInt(reqRoleSorts)

	user := model.User{
		Model:    oldUser.Model,
		Name:     req.Username,
		Password: oldUser.Password,
		//Mobile:       req.Mobile,
		Email:        req.Email,
		Avatar:       req.Avatar,
		Nickname:     req.Nickname,
		Introduction: req.Introduction,
		Status:       req.Status,
		Creator:      ctxUser.Name,
		Roles:        roles,
	}
	// Determining whether to update yourself or someone else
	if userId == int(ctxUser.ID) {
		// If you are updating yourself
		// cannot disable itself
		if req.Status == 2 {
			response.Fail(c, nil, "Cannot disable yourself")
			return
		}
		// Cannot change your role
		reqDiff, currentDiff := funk.Difference(req.RoleIds, currentRoleIds)
		if len(reqDiff.([]uint)) > 0 || len(currentDiff.([]uint)) > 0 {
			response.Fail(c, nil, "Cannot change your own role")
			return
		}

		// You cannot update your own password, only in your personal centre
		if req.Password != "" {
			response.Fail(c, nil, "Please go to the personal center to update your own password")
			return
		}

		// Password assignment
		user.Password = ctxUser.Password

	} else {
		// If updating someone else
		// A user cannot update a user with a higher role rank than their own or the same rank
		// Get the user role sort minimum based on the userIdID in the path
		minRoleSorts, err := uc.UserRepository.GetUserMinRoleSortsByIds([]uint{uint(userId)})
		if err != nil {
			response.Fail(c, nil, "Failed to get the minimum user role sort value by user ID")
			return
		}
		if len(minRoleSorts) > 0 && currentRoleSortMin >= minRoleSorts[0] {
			response.Fail(c, nil, "Users cannot update users with higher or equal levels than themselves")
			return
		}

		// Users cannot update another user's role level to be higher or equal to their own
		if currentRoleSortMin >= reqRoleSortMin {
			response.Fail(c, nil, "Users cannot update other users' role levels to be higher or equal to themselves")
			return
		}

		// Password assignment
		if req.Password != "" {
			// Password decryption via RSA
			decodeData, err := util.RSADecrypt([]byte(req.Password), config.Conf.System.PrivateKey)
			if err != nil {
				response.Fail(c, nil, "Failed to operate password")
				log.Log.Error(err)
				return
			}
			req.Password = string(decodeData)
			user.Password = util.GenPasswd(req.Password)
		}

	}

	// Update user
	err = uc.UserRepository.UpdateUser(&user)
	if err != nil {
		response.Fail(c, nil, "Failed to update user")
		log.Log.Errorf("update user error: %v", err)
		return
	}
	response.Success(c, nil, "Successfully updated user")
}

// Batch delete users
func (uc UserController) BatchDeleteUserByIds(c *gin.Context) {
	var req vo.DeleteUserRequest
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

	// User ID passed from the front end
	reqUserIds := req.UserIds

	// Get the minimum value of user role sorting based on user ID
	roleMinSortList, err := uc.UserRepository.GetUserMinRoleSortsByIds(reqUserIds)
	if err != nil {
		response.Fail(c, nil, "Failed to get the minimum user role sort value by user ID")
		return
	}

	// Current user role sort minimum (highest ranked role) and current user
	minSort, ctxUser, err := uc.UserRepository.GetCurrentUserMinRoleSort(c)
	if err != nil {
		response.Fail(c, nil, "Failed to get current user")
		log.Log.Errorf("get current user error: %v", err)
		return
	}
	currentRoleSortMin := int(minSort)

	// Cannot delete yourself
	if funk.Contains(reqUserIds, ctxUser.ID) {
		response.Fail(c, nil, "Cannot delete yourself")
		return
	}

	// You cannot delete a user with a lower rank (higher rank) than your own character
	for _, sort := range roleMinSortList {
		if currentRoleSortMin >= sort {
			response.Fail(c, nil, "Users cannot delete users with higher role levels than themselves")
			return
		}
	}

	for _, id := range reqUserIds {
		user, err := uc.UserRepository.GetUserById(id)
		if err != nil {
			response.Fail(c, nil, "Failed to delete user")
			log.Log.Errorf("Failed to delete user: %v", err)
			return
		}
		err = uc.HeadscaleUserRepository.DeleteUserWithString(user.Name)
		if err != nil {
			response.Fail(c, nil, "Failed to delete user:"+user.Name)
			log.Log.Errorf("delete user error: %v, %v", user.Name, err)
			return
		}
	}
	response.Success(c, nil, "Successfully deleted user")
}
