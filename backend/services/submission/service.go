package submission

import (
	"errors"
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/models"
	"sample-exchange/backend/services/samplepack"

	"gorm.io/gorm"
)

var (
	ErrSubmissionClosed  = errors.New("submission window is closed")
)

type Service struct {
	config *config.Config
	db     *gorm.DB
	packService *samplepack.Service
}

func NewService(cfg *config.Config, packService *samplepack.Service) *Service {
	return &Service{
		config: cfg,
		db:     db.DB,
		packService: packService,
	}
}

func (s *Service) CreateSubmission(userID uint, submission *models.Submission) error {
	if !s.packService.IsSubmissionAllowed() {
		return ErrSubmissionClosed
	}

	currentPack, err := s.packService.GetCurrentPack()
	if err != nil {
		return err
	}

	submission.UserID = userID
	submission.SamplePackID = currentPack.ID
	submission.SubmittedAt = time.Now()

	return s.db.Create(submission).Error
}

func (s *Service) GetSubmission(id uint) (*models.Submission, error) {
	var submission models.Submission
	err := s.db.Preload("User").
		Preload("SamplePack").
		First(&submission, id).Error

	if err == gorm.ErrRecordNotFound {
		return nil, models.ErrSubmissionNotFound
	}

	if err != nil {
		return nil, err
	}

	return &submission, nil
}

func (s *Service) ListSubmissions(packID uint, limit, offset int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.db.Where("sample_pack_id = ?", packID).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}

	return submissions, nil
} 