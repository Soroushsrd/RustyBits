package models

import "time"

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;not null"`
	Content   string    `json:"content" gorm:"type:text"`
	Excerpt   string    `json:"excerpt" `
	Published bool      `json:"published" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []Tag     `json:"tags" gorm:"many2many:post_tags;"`
}

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"`
}

type Tag struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"uniqueIndex;not null"`
	Posts []Post `json:"-" gorm:"many2many:post_tags;"`
}
