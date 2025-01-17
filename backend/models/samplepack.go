package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Error definitions
var (
	ErrPackNotFound = errors.New("sample pack not found")
)

type SamplePack struct {
	gorm.Model
	Title       string       `json:"title" gorm:"not null"`
	Description string       `json:"description"`
	StartDate   time.Time    `json:"startDate"`
	EndDate     time.Time    `json:"endDate"`
	UploadStart time.Time    `json:"uploadStart"`
	UploadEnd   time.Time    `json:"uploadEnd"`
	IsActive    bool         `json:"isActive" gorm:"default:false"`
	Samples     []Sample     `json:"samples"`
	Submissions []Submission `json:"submissions,omitempty"`
}
