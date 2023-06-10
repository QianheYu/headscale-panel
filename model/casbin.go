package model

// 角色权限规则
type RoleCasbin struct {
	Keyword string `json:"keyword"` // Character Keywords
	Path    string `json:"path"`    // Access path
	Method  string `json:"method"`  // Request method
}
