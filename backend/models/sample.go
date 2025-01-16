package models

import (
	"time"

	"gorm.io/gorm"
)

type Sample struct {
	gorm.Model
	Filename     string    `json:"filename" gorm:"not null"`
	FileSize     int64     `json:"fileSize"`
	FilePath     string    `json:"-"` // Internal storage path
	FileURL      string    `json:"fileUrl"`
	UploadedAt   time.Time `json:"uploadedAt"`
	UserID       uint      `json:"userId"`
	User         User      `json:"user"`
	SamplePackID uint      `json:"samplePackId"`
	SamplePack   SamplePack `json:"samplePack" gorm:"foreignKey:SamplePackID"`
} 