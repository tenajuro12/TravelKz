package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Email    string `gorm:"unique;	not null"`
	Password string `gorm:"not null"`
	IsAdmin  bool   `json:"is_admin" gorm:"default:false"`
}
