package submission

import (
	"errors"
	"fmt"
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/models"
	"sample-exchange/backend/services/samplepack"

	"gorm.io/gorm"
)

var (
	ErrSubmissionClosed = errors.New("submission window is closed")
)

type Service struct {
	config      *config.Config
	packService *samplepack.Service
}

func NewService(cfg *config.Config, packService *samplepack.Service) *Service {
	return &Service{
		config:      cfg,
		packService: packService,
	}
}

func (s *Service) CreateSubmission(userID uint, submission *models.Submission) error {
	if !s.packService.IsSubmissionAllowed() {
		pack, err := s.packService.GetCurrentPack()
		if err != nil {
			return fmt.Errorf("no active sample pack found")
		}
		return fmt.Errorf("submission window is closed. Opens %s, closes %s",
			pack.StartDate.Format("Jan 2 15:04 MST"),
			pack.EndDate.Format("Jan 2 15:04 MST"))
	}

	currentPack, err := s.packService.GetCurrentPack()
	if err != nil {
		return err
	}

	submission.UserID = userID
	submission.SamplePackID = currentPack.ID
	submission.SubmittedAt = time.Now()

	return db.GetDB().Create(submission).Error
}

func (s *Service) GetSubmission(id uint) (*models.Submission, error) {
	var submission models.Submission
	err := db.GetDB().Preload("User").
		Preload("SamplePack").
		First(&submission, id).Error

	if err == gorm.ErrRecordNotFound {
		return nil, models.ErrSubmissionNotFound
	}

	if err != nil {
		return nil, err
	}

	submission.FileURL = fmt.Sprintf("/api/submissions/%d/download", submission.ID)

	return &submission, nil
}

// CreateTestSubmission creates a test submission for development
func (s *Service) CreateTestSubmission(userID uint, packID uint) (*models.Submission, error) {
	submission := &models.Submission{
		Title:        "Test Submission",
		Description:  "A submission for testing",
		Filename:     "test_submission.mp3",
		FilePath:     "/tmp/test_submission.mp3",
		FileSize:     1024,
		UserID:       userID,
		SamplePackID: packID,
		SubmittedAt:  time.Now(),
	}

	if err := db.GetDB().Create(submission).Error; err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *Service) ListSubmissions(packID uint, limit, offset int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := db.GetDB().Where("sample_pack_id = ?", packID).
		Preload("User").
		Preload("SamplePack").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}

	// Generate file URLs for submissions
	for i := range submissions {
		submissions[i].FileURL = fmt.Sprintf("/api/submissions/%d/download", submissions[i].ID)
	}

	return submissions, nil
}
