package models

import (
	"time"

	"gorm.io/gorm"
)

type Sample struct {
    gorm.Model
    Filename    string    `json:"filename"`
    UploadedBy  uint      `json:"uploaded_by"`
    UploadedAt  time.Time `json:"uploaded_at"`
    FileSize    int64     `json:"file_size"`
    Duration    float64   `json:"duration"`
    FilePath    string    `json:"-"` // Internal storage path
}

type SamplePack struct {
    gorm.Model
    StartDate   time.Time `json:"start_date"`
    EndDate     time.Time `json:"end_date"`
    IsActive    bool      `json:"is_active"`
    Samples     []Sample  `json:"samples" gorm:"foreignKey:SamplePackID"`
} 