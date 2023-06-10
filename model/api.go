package model

import "gorm.io/gorm"

type Api struct {
	gorm.Model
	Method   string `gorm:"type:varchar(20);comment:Request method" json:"method"`
	Path     string `gorm:"type:varchar(100);comment:Access path" json:"path"`
	Category string `gorm:"type:varchar(50);comment:Category" json:"category"`
	Desc     string `gorm:"type:varchar(100);comment:Description" json:"desc"`
	Creator  string `gorm:"type:varchar(20);comment:Created by" json:"creator"`
}
