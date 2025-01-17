package models

import (
	"time"

	"gorm.io/gorm"
)

type SamplePack struct {
	ID          uint           `json:"ID" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	UploadStart time.Time      `json:"uploadStart"`
	UploadEnd   time.Time      `json:"uploadEnd"`
	StartDate   time.Time      `json:"startDate"`
	EndDate     time.Time      `json:"endDate"`
	IsActive    bool           `json:"isActive" gorm:"default:false"`
	Samples     []Sample       `json:"samples"`
	Submissions []Submission   `json:"submissions"`
}
