package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name         string  `gorm:"type:varchar(63);not null;unique" json:"username"`
	Password     string  `gorm:"size:255" json:"password"`
	Email        string  `gorm:"type:varchar(255);unique" json:"email"`
	Avatar       string  `gorm:"type:varchar(255)" json:"avatar"`
	Nickname     string  `gorm:"type:varchar(20)" json:"nickname"`
	Introduction string  `gorm:"type:varchar(255)" json:"introduction"`
	Status       uint    `gorm:"type:smallint;default:1;comment:1 normal, 2 disabled" json:"status"`
	Creator      string  `gorm:"type:varchar(20);" json:"creator"`
	Roles        []*Role `gorm:"many2many:user_roles" json:"roles"`
	RefreshFlag  bool    `gorm:"-" json:"-"`
}
