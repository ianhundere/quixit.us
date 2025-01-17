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
	gorm.Model
	Title        string     `json:"title" gorm:"not null"`
	Description  string     `json:"description"`
	FileURL      string     `json:"fileUrl"`
	FilePath     string     `json:"-"` // Internal storage path
	FileSize     int64      `json:"fileSize"`
	UserID       uint       `json:"userId"`
	User         User       `json:"user" gorm:"foreignKey:UserID"`
	SamplePackID uint       `json:"samplePackId"`
	SamplePack   SamplePack `json:"samplePack" gorm:"foreignKey:SamplePackID"`
	SubmittedAt  time.Time  `json:"submittedAt"`
}
