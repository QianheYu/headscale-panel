package vo

// New role structure
type CreateRoleRequest struct {
	Name    string `json:"name" form:"name" validate:"required,min=1,max=20"`
	Keyword string `json:"keyword" form:"keyword" validate:"required,min=1,max=20"`
	Desc    string `json:"desc" form:"desc" validate:"min=0,max=100"`
	Status  uint   `json:"status" form:"status" validate:"oneof=1 2"`
	Sort    uint   `json:"sort" form:"sort" validate:"gte=1,lte=999"`
}

// Get the user role structure
type RoleListRequest struct {
	Name     string `json:"name" form:"name"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   uint   `json:"status" form:"status"`
	PageNum  uint   `json:"pageNum" form:"pageNum"`
	PageSize uint   `json:"pageSize" form:"pageSize"`
}

// Bulk Deletion of Role Structs
type DeleteRoleRequest struct {
	RoleIds []uint `json:"roleIds" form:"roleIds"`
}

// Update the role's permissions menu
type UpdateRoleMenusRequest struct {
	MenuIds []uint `json:"menuIds" form:"menuIds"`
}

// Permissions interface for updating roles
type UpdateRoleApisRequest struct {
	ApiIds []uint `json:"apiIds" form:"apiIds"`
}
