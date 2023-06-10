package model

import "gorm.io/gorm"

type Menu struct {
	gorm.Model
	Name       string  `gorm:"type:varchar(50);comment:Menu name" json:"name"`
	Title      string  `gorm:"type:varchar(50);comment:Menu title" json:"title"`
	Icon       *string `gorm:"type:varchar(50);comment:Menu icons" json:"icon"`
	Path       string  `gorm:"type:varchar(100);comment:Menu access path" json:"path"`
	Redirect   *string `gorm:"type:varchar(100);comment:Redirect path" json:"redirect"`
	Component  string  `gorm:"type:varchar(100);comment:Front-end component path" json:"component"`
	Sort       uint    `gorm:"type:int;default:999;comment:Menu order (1-999)" json:"sort"`
	Status     uint    `gorm:"type:smallint;default:1;comment:Menu status (normal/disabled, default normal)" json:"status"`
	Hidden     uint    `gorm:"type:smallint;default:2;comment:Menu hidden in sidebar (1 hide, 2 show)" json:"hidden"`
	Cache      uint    `gorm:"type:smallint;default:2;comment:Whether the menu is <keep-alive cached (1 not cached, 2 cached)" json:"cache"`
	AlwaysShow uint    `gorm:"type:smallint;default:2;comment:Ignore previously defined rules and always show the root route (1 ignore, 2 don\'t ignore)" json:"alwaysShow"`
	Breadcrumb uint    `gorm:"type:smallint;default:1;comment:Breadcrumb visibility (1 visible/2 hidden, visible by default)" json:"breadcrumb"`
	ActiveMenu *string `gorm:"type:varchar(100);comment:Routing that you want to highlight in the sidebar when it routes" json:"activeMenu"`
	ParentId   uint    `gorm:"default:0;comment:Parent menu number (a number of 0 indicates the root menu)" json:"parentId"`
	Creator    string  `gorm:"type:varchar(20);comment:Created by" json:"creator"`
	Children   []*Menu `gorm:"-" json:"children"`                  // Sub-menu collection
	Roles      []*Role `gorm:"many2many:role_menus;" json:"roles"` // Role menu many-to-many relationships
}
