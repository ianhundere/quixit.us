package models

import (
	"time"

	"gorm.io/gorm"
)

type Sample struct {
	ID           uint           `json:"ID" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Filename     string         `json:"filename"`
	FileURL      string         `json:"fileUrl" gorm:"-"`
	FilePath     string         `json:"-"`
	FileSize     int64          `json:"fileSize"`
	UserID       uint           `json:"userID"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	SamplePackID uint           `json:"samplePackID"`
	SamplePack   SamplePack     `json:"samplePack" gorm:"foreignKey:SamplePackID"`
}
