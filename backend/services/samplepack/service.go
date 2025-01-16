package samplepack

import (
	"log"
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
	log.Printf("Fetching current pack...")
	
	err := s.db.Where("is_active = ?", true).
		Preload("Samples").
		First(&pack).Error

	if err == gorm.ErrRecordNotFound {
		log.Printf("No active pack found")
		return nil, nil
	}
	
	log.Printf("Found active pack: %+v", pack)
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
	log.Printf("Listing packs with limit: %d", limit)
	
	err := s.db.Order("created_at DESC").
		Limit(limit).
		Preload("Samples").
		Find(&packs).Error
		
	log.Printf("Found %d packs", len(packs))
	return packs, err
}

func (s *Service) CreatePack() (*models.SamplePack, error) {
	// Deactivate any currently active packs
	s.db.Model(&models.SamplePack{}).Where("is_active = ?", true).Update("is_active", false)

	// Calculate time windows
	now := time.Now()
	
	// Find next Friday 00:00
	uploadStart := nextWeekday(now, time.Friday)
	uploadStart = time.Date(uploadStart.Year(), uploadStart.Month(), uploadStart.Day(), 0, 0, 0, 0, uploadStart.Location())
	
	// Sunday 23:59:59
	uploadEnd := uploadStart.Add(72 * time.Hour).Add(-1 * time.Second)
	
	// Next Monday 00:00
	startDate := nextWeekday(uploadEnd, time.Monday)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	
	// Friday 00:00
	endDate := startDate.Add(96 * time.Hour)

	pack := &models.SamplePack{
		Title:       "",  // Will be set by caller
		Description: "",  // Will be set by caller
		UploadStart: uploadStart,
		UploadEnd:   uploadEnd,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    true,
	}

	return pack, s.db.Create(pack).Error
}

// Helper function to find next occurrence of a weekday
func nextWeekday(t time.Time, weekday time.Weekday) time.Time {
	daysUntil := int(weekday - t.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return t.AddDate(0, 0, daysUntil)
}

func (s *Service) IsUploadAllowed() bool {
	if s.config.BypassTimeWindows {
		return true
	}

	now := time.Now()
	pack, err := s.GetCurrentPack()
	if err != nil {
		return false
	}
	log.Printf("Upload window: %s to %s (now: %s)", 
		pack.UploadStart.Format(time.RFC3339),
		pack.UploadEnd.Format(time.RFC3339),
		now.Format(time.RFC3339))
	return now.After(pack.UploadStart) && now.Before(pack.UploadEnd)
}

func (s *Service) IsSubmissionAllowed() bool {
	if s.config.BypassTimeWindows {
		return true
	}

	now := time.Now()
	pack, err := s.GetCurrentPack()
	if err != nil {
		return false
	}
	log.Printf("Submission window: %s to %s (now: %s)", 
		pack.StartDate.Format(time.RFC3339),
		pack.EndDate.Format(time.RFC3339),
		now.Format(time.RFC3339))
	return now.After(pack.StartDate) && now.Before(pack.EndDate)
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