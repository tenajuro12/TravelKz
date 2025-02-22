package models

import "gorm.io/gorm"

type Blog struct {
	gorm.Model
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	UserID    uint       `json:"user_id"`
	Likes     int        `json:"likes"`
	Category  string     `json:"category"`
	Comments  []Comment  `gorm:"foreignKey:BlogID" json:"comments"`
	BlogLikes []BlogLike `gorm:"foreignKey:BlogID" json:"blog_likes"`
}

type Comment struct {
	gorm.Model
	Content string `json:"content"`
	BlogID  uint   `json:"blog_id"`
	UserID  uint   `json:"user_id"`
	Blog    Blog   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"blog"`
}

type BlogLike struct {
	gorm.Model
	UserID uint `json:"user_id"`
	BlogID uint `json:"blog_id"`
	Blog   Blog `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"blog"`
}
