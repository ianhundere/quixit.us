package models

import (
	"time"

	"gorm.io/gorm"
)

// SamplePack represents a collection of audio samples and their submissions.
// Core Features:
// 1. Time Windows:
//    a) Upload Window (72 hours):
//       - Starts: Friday 00:00
//       - Ends: Sunday 23:59:59
//       - Users upload individual audio samples
//    b) Submission Window (12 days):
//       - Starts: Monday 00:00
//       - Ends: Following Friday 23:59:59
//       - Users submit songs made from the samples
//
// 2. Sample Management:
//    - Users can upload audio samples during upload window
//    - Samples are collected into a downloadable pack
//    - Pack becomes available at start of submission window
//
// 3. Submission Rules:
//    - Songs must only use samples from the current pack
//    - Submissions accepted only during submission window
//    - Multiple submissions allowed per user
//
// 4. User Authentication:
//    - OAuth-based user accounts required for all actions
//    - Supported providers: GitHub, Google, Discord
//    - User profiles track all submissions and samples
//    - Authentication required for:
//      * Uploading samples
//      * Downloading sample packs
//      * Submitting songs
//      * Viewing submission history
//
// 5. Sample Upload Rules:
//    - Upload window: Friday 00:00 to Sunday 23:59:59 (72 hours)
//    - Supported formats: .wav, .mp3, .aiff, .flac
//    - Maximum file size: 50MB per sample
//    - One sample per upload
//    - Multiple uploads allowed per user
//    - No duplicate filenames allowed within a pack
//    - Samples become available to all users when pack opens
type SamplePack struct {
	ID          uint           `json:"ID" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	UploadStart time.Time      `json:"uploadStart"`    // Friday 00:00
	UploadEnd   time.Time      `json:"uploadEnd"`      // Sunday 23:59:59
	StartDate   time.Time      `json:"startDate"`      // Monday 00:00
	EndDate     time.Time      `json:"endDate"`        // Friday 23:59:59 (12 days later)
	IsActive    bool           `json:"isActive" gorm:"default:false"`
	Samples     []Sample       `json:"samples"`
	Submissions []Submission   `json:"submissions"`
}
