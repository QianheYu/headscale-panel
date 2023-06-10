package dto

import "headscale-panel/model"

// Current user information returned to the front end
type UserInfoDto struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	//Mobile       string        `json:"mobile"`
	Email        string        `json:"email"`
	Avatar       string        `json:"avatar"`
	Nickname     string        `json:"nickname"`
	Introduction string        `json:"introduction"`
	Roles        []*model.Role `json:"roles"`
}

func ToUserInfoDto(user model.User) UserInfoDto {
	return UserInfoDto{
		ID:       user.ID,
		Username: user.Name,
		//Mobile:       user.Mobile,
		Email:        user.Email,
		Avatar:       user.Avatar,
		Nickname:     user.Nickname,
		Introduction: user.Introduction,
		Roles:        user.Roles,
	}
}

// List of users returned to the front end
type UsersDto struct {
	ID       uint   `json:"ID"`
	Username string `json:"username"`
	//Mobile       string `json:"mobile"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	Nickname     string `json:"nickname"`
	Introduction string `json:"introduction"`
	Status       uint   `json:"status"`
	Creator      string `json:"creator"`
	RoleIds      []uint `json:"roleIds"`
}

func ToUsersDto(userList []*model.User) []UsersDto {
	var users []UsersDto
	for _, user := range userList {
		userDto := UsersDto{
			ID:       user.ID,
			Username: user.Name,
			//Mobile:       user.Mobile,
			Email:        user.Email,
			Avatar:       user.Avatar,
			Nickname:     user.Nickname,
			Introduction: user.Introduction,
			Status:       user.Status,
			Creator:      user.Creator,
		}
		roleIds := make([]uint, 0)
		for _, role := range user.Roles {
			roleIds = append(roleIds, role.ID)
		}
		userDto.RoleIds = roleIds
		users = append(users, userDto)
	}

	return users
}
