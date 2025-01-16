package models

import (
	"time"

	"gorm.io/gorm"
)

type Submission struct {
    gorm.Model
    Title       string    `json:"title"`
    Artist      string    `json:"artist"`
    SubmittedBy uint      `json:"submitted_by"`
    SubmittedAt time.Time `json:"submitted_at"`
    FilePath    string    `json:"-"`
    SamplePackID uint     `json:"sample_pack_id"`
} 