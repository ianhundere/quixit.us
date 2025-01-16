package samplepack

import (
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/errors"
	"sample-exchange/backend/models"

	"gorm.io/gorm"
)

type Service struct {
	config *config.Config
	db     *gorm.DB
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
		db:     db.DB,
	}
}

func (s *Service) GetCurrentPack() (*models.SamplePack, error) {
	var pack models.SamplePack
	err := s.db.Where("is_active = ?", true).
		Preload("Samples").
		First(&pack).Error

	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewNotFoundError("Active sample pack")
	}
	return &pack, err
}

func (s *Service) GetPack(id uint) (*models.SamplePack, error) {
	var pack models.SamplePack
	err := s.db.Preload("Samples").First(&pack, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewNotFoundError("Sample pack")
	}
	return &pack, err
}

func (s *Service) ListPacks(limit int) ([]models.SamplePack, error) {
	var packs []models.SamplePack
	err := s.db.Order("created_at DESC").
		Limit(limit).
		Preload("Samples").
		Find(&packs).Error
	return packs, err
}

func (s *Service) CreatePack() (*models.SamplePack, error) {
	// Deactivate current active pack if exists
	s.db.Model(&models.SamplePack{}).
		Where("is_active = ?", true).
		Update("is_active", false)

	// Create new pack
	pack := &models.SamplePack{
		StartDate: time.Now(),
		EndDate:   time.Now().Add(s.config.UploadDuration),
		IsActive:  true,
	}

	err := s.db.Create(pack).Error
	return pack, err
}

func (s *Service) IsUploadAllowed() bool {
	now := time.Now()
	return now.Weekday() == s.config.UploadStartDay ||
		now.Before(time.Now().Add(s.config.UploadDuration))
}

func (s *Service) IsSubmissionAllowed() bool {
	now := time.Now()
	return now.Weekday() == s.config.SubmissionStartDay ||
		now.Before(time.Now().Add(s.config.SubmissionDuration))
}

func (s *Service) AddSample(packID uint, sample *models.Sample) error {
	pack, err := s.GetPack(packID)
	if err != nil {
		return err
	}

	if !s.IsUploadAllowed() {
		return errors.NewAuthorizationError("Upload window is closed")
	}

	return s.db.Model(pack).Association("Samples").Append(sample)
} 