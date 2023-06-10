package vo

// Get the interface list structure
type ApiListRequest struct {
	Method   string `json:"method" form:"method"`
	Path     string `json:"path" form:"path"`
	Category string `json:"category" form:"category"`
	Creator  string `json:"creator" form:"creator"`
	PageNum  uint   `json:"pageNum" form:"pageNum"`
	PageSize uint   `json:"pageSize" form:"pageSize"`
}

// Creating Interface Structures
type CreateApiRequest struct {
	Method   string `json:"method" form:"method" validate:"required,min=1,max=20"`
	Path     string `json:"path" form:"path" validate:"required,min=1,max=100"`
	Category string `json:"category" form:"category" validate:"required,min=1,max=50"`
	Desc     string `json:"desc" form:"desc" validate:"min=0,max=100"`
}

// Update the interface structure
type UpdateApiRequest struct {
	Method   string `json:"method" form:"method" validate:"min=1,max=20"`
	Path     string `json:"path" form:"path" validate:"min=1,max=100"`
	Category string `json:"category" form:"category" validate:"min=1,max=50"`
	Desc     string `json:"desc" form:"desc" validate:"min=0,max=100"`
}

// Bulk Deletion of Interface Structs
type DeleteApiRequest struct {
	ApiIds []uint `json:"apiIds" form:"apiIds"`
}
