package vo

// User Login Structs
type RegisterAndLoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// Creating User Structures
type CreateUserRequest struct {
	Username string `form:"username" json:"username" validate:"required,lowercase,min=2,max=63"`
	Password string `form:"password" json:"password"`
	//Mobile       string `form:"mobile" json:"mobile" validate:"required,checkMobile"`
	Email        string `form:"email" json:"email" validate:"required,email"`
	Avatar       string `form:"avatar" json:"avatar" validate:"min=0,max=150,omitempty,url"`
	Nickname     string `form:"nickname" json:"nickname" validate:"min=0,max=20"`
	Introduction string `form:"introduction" json:"introduction" validate:"min=0,max=255"`
	Status       uint   `form:"status" json:"status" validate:"oneof=1 2"`
	RoleIds      []uint `form:"roleIds" json:"roleIds" validate:"required"`
}

// Get the user list structure
type UserListRequest struct {
	Username string `json:"username" form:"username" `
	//Mobile   string `json:"mobile" form:"mobile" `
	Email    string `json:"email" form:"email" `
	Nickname string `json:"nickname" form:"nickname" `
	Status   uint   `json:"status" form:"status" `
	PageNum  uint   `json:"pageNum" form:"pageNum"`
	PageSize uint   `json:"pageSize" form:"pageSize"`
}

// Bulk Deletion of User Structs
type DeleteUserRequest struct {
	UserIds []uint `json:"userIds" form:"userIds"`
}

// Update the password structure
type ChangePwdRequest struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" validate:"required"`
}
