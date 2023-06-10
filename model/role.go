package model

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name    string  `gorm:"type:varchar(20);not null;unique" json:"name"`
	Keyword string  `gorm:"type:varchar(20);not null;unique" json:"keyword"`
	Desc    *string `gorm:"type:varchar(100);" json:"desc"`
	Home    string  `gorm:"type:varchar(20);not null" json:"home"`
	Status  uint    `gorm:"type:smallint;default:1;comment:1 normal, 2 disabled" json:"status"`
	Sort    uint    `gorm:"type:smallint;default:999;comment:Role sorting (the higher the sort the lower the permissions, can\'t view roles with smaller serial numbers than yourself, can\'t edit user permissions with the same serial number, sorted 1 means super admin)" json:"sort"`
	Creator string  `gorm:"type:varchar(20);" json:"creator"`
	Users   []*User `gorm:"many2many:user_roles" json:"users"`
	Menus   []*Menu `gorm:"many2many:role_menus;" json:"menus"` // 角色菜单多对多关系
}
