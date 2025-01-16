package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidSubmission = errors.New("invalid submission")
	ErrAlreadyVoted     = errors.New("already voted for this submission")
	ErrSubmissionClosed    = errors.New("submission is closed")
	ErrSubmissionNotFound  = errors.New("submission not found")
)

type Submission struct {
	gorm.Model
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	FileURL     string    `json:"fileUrl"`
	FilePath    string    `json:"-"` // Internal storage path
	FileSize    int64     `json:"fileSize"`
	UserID      uint      `json:"userId"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	SamplePackID uint     `json:"samplePackId"`
	SamplePack  SamplePack `json:"samplePack" gorm:"foreignKey:SamplePackID"`
	SubmittedAt time.Time `json:"submittedAt"`
}

type Vote struct {
	gorm.Model
	UserID       uint       `json:"userId"`
	User         User       `json:"user"`
	SubmissionID uint       `json:"submissionId"`
	Submission   Submission `json:"-"`
	Value        int        `json:"value" gorm:"check:value IN (-1, 1)"` // -1 for downvote, 1 for upvote
} 