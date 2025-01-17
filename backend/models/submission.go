package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidSubmission  = errors.New("invalid submission")
	ErrSubmissionClosed   = errors.New("submission is closed")
	ErrSubmissionNotFound = errors.New("submission not found")
)

type Submission struct {
	ID           uint           `json:"ID" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Filename     string         `json:"filename"`
	FileURL      string         `json:"fileUrl" gorm:"-"`
	FilePath     string         `json:"-"`
	FileSize     int64          `json:"fileSize"`
	UserID       uint           `json:"userID"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	SamplePackID uint           `json:"samplePackID"`
	SamplePack   SamplePack     `json:"samplePack" gorm:"foreignKey:SamplePackID"`
	SubmittedAt  time.Time      `json:"submittedAt"`
}
