package model

import (
	"gorm.io/gorm"
	"time"
)

type OperationLog struct {
	gorm.Model
	Username   string    `gorm:"type:varchar(20);comment:User Login Name" json:"username"`
	Ip         string    `gorm:"type:inet;comment:IP address" json:"ip"`
	IpLocation string    `gorm:"type:varchar(20);comment:IP Location" json:"ipLocation"`
	Method     string    `gorm:"type:varchar(20);comment:Request method" json:"method"`
	Path       string    `gorm:"type:varchar(100);comment:Access path" json:"path"`
	Desc       string    `gorm:"type:varchar(100);comment:Description" json:"desc"`
	Status     int       `gorm:"type:int4;comment:Response Status Code" json:"status"`
	StartTime  time.Time `gorm:"type:timestamp(3);comment:Initiation time" json:"startTime"`
	TimeCost   int64     `gorm:"type:int8;comment:Time taken to request(ms)" json:"timeCost"`
	UserAgent  string    `gorm:"type:varchar(20);comment:Browser identification" json:"userAgent"`
}
