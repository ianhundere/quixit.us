package models

import (
	"time"

	"gorm.io/gorm"
)

type Sample struct {
	gorm.Model
	Filename     string     `json:"filename" gorm:"not null"`
	FileSize     int64      `json:"fileSize"`
	FilePath     string     `json:"-"`       // Internal storage path
	FileURL      string     `json:"fileUrl"` // Public URL for download
	UploadedAt   time.Time  `json:"uploadedAt"`
	UserID       uint       `json:"userId"`
	User         User       `json:"user" gorm:"foreignKey:UserID"`
	SamplePackID uint       `json:"samplePackId"`
	SamplePack   SamplePack `json:"-" gorm:"foreignKey:SamplePackID"` // Hide in JSON to avoid recursion
}
