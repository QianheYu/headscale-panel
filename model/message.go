/* This part of the code is reserved for functionality */
package model

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	Title    string `gorm:"type:varchar(50);not null" json:"title"`
	Content  string `gorm:"type:text;not null" json:"content"`
	Type     uint   `gorm:"type:smallint;default:1;comment:Message category (0-255)" json:"type"`
	Link     string `gorm:"type:varchar(50)" json:"link"`
	Creator  string `gorm:"type:varchar(10)" json:"creator"`
	HaveRead bool   `gorm:"type:boolean;not null" json:"have_read"`
}
