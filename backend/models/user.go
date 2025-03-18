package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"ID" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Email    string `json:"email" gorm:"unique;not null"`
	Name     string `json:"name"`
	Provider string `json:"provider"` // OAuth provider (github, google, discord)
	Avatar   string `json:"avatar"`   // URL to user's avatar
	IsAdmin  bool   `json:"isAdmin" gorm:"default:false"` // Whether the user has admin privileges
}
