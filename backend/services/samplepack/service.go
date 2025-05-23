package samplepack

import (
	"archive/zip"
	stderrors "errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/errors"
	"sample-exchange/backend/models"

	"gorm.io/gorm"
)

type Service struct {
	cfg *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) GetCurrentPack() (*models.SamplePack, error) {
	var pack models.SamplePack
	result := db.GetDB().Where("is_active = ?", true).First(&pack)
	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &pack, nil
}

func (s *Service) GetPack(id uint) (*models.SamplePack, error) {
	var pack models.SamplePack
	err := db.GetDB().
		Preload("Samples").
		Preload("Samples.User").
		First(&pack, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewNotFoundError("Sample pack")
	}
	return &pack, err
}

func (s *Service) ListPacks(limit int) ([]models.SamplePack, error) {
	var packs []models.SamplePack
	result := db.GetDB().Where("is_active = ?", false).Order("created_at desc").Limit(limit).Find(&packs)
	if result.Error != nil {
		return nil, result.Error
	}
	return packs, nil
}

func (s *Service) CreatePack() (*models.SamplePack, error) {
	// Deactivate any currently active pack
	if err := db.GetDB().Model(&models.SamplePack{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		return nil, err
	}

	// Set time windows
	now := time.Now()
	uploadStart := now
	uploadEnd := now.Add(72 * time.Hour)    // 3 days for upload window
	startDate := uploadEnd
	endDate := startDate.Add(240 * time.Hour) // 10 days for submission window

	pack := &models.SamplePack{
		Title:       "", // Will be set by caller
		Description: "", // Will be set by caller
		UploadStart: uploadStart,
		UploadEnd:   uploadEnd,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    true,
	}

	return pack, db.GetDB().Create(pack).Error
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
	if s.cfg.BypassTimeWindows {
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
	if s.cfg.BypassTimeWindows {
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

	if !s.IsUploadAllowedForPack(packID) {
		return errors.NewAuthorizationError("Upload window is closed")
	}

	return db.GetDB().Model(pack).Association("Samples").Append(sample)
}

// CreatePackZip creates a zip file containing all samples in a pack
// CreateTestPack creates a sample pack with test data
func (s *Service) CreateTestPack(userID uint) (*models.SamplePack, error) {
	pack, err := s.CreatePack()
	if err != nil {
		return nil, err
	}

	pack.Title = "Test Sample Pack"
	pack.Description = "A sample pack for testing"
	
	// Set time windows to be currently active
	now := time.Now()
	pack.UploadStart = now.Add(-24 * time.Hour)  // Started yesterday
	pack.UploadEnd = now.Add(24 * time.Hour)     // Ends tomorrow
	pack.StartDate = now.Add(-24 * time.Hour)    // Started yesterday
	pack.EndDate = now.Add(7 * 24 * time.Hour)   // Ends in a week

	if err := db.GetDB().Save(pack).Error; err != nil {
		return nil, err
	}

	// Create a test sample
	sample := &models.Sample{
		Filename:     "test_sample.wav",
		FilePath:     "/tmp/test_sample.wav",
		FileSize:     1024,
		UserID:       userID,
		SamplePackID: pack.ID,
	}

	if err := s.AddSample(pack.ID, sample); err != nil {
		return nil, err
	}

	return pack, nil
}

func (s *Service) CreatePackZip(pack models.SamplePack, zipPath string) error {
	log.Printf("Creating zip file for pack %d with %d samples", pack.ID, len(pack.Samples))

	// Create the zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		log.Printf("Failed to create zip file at %s: %v", zipPath, err)
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add each sample to the zip
	for _, sample := range pack.Samples {
		log.Printf("Adding sample %d (%s) from %s", sample.ID, sample.Filename, sample.FilePath)

		// Open the sample file using the stored file path
		sampleFile, err := os.Open(sample.FilePath)
		if err != nil {
			log.Printf("Failed to open sample file %d at %s: %v", sample.ID, sample.FilePath, err)
			return fmt.Errorf("failed to open sample file %d: %w", sample.ID, err)
		}
		defer sampleFile.Close()

		// Create a new file in the zip
		zipEntry, err := zipWriter.Create(sample.Filename)
		if err != nil {
			log.Printf("Failed to create zip entry for sample %d: %v", sample.ID, err)
			return fmt.Errorf("failed to create zip entry for sample %d: %w", sample.ID, err)
		}

		// Copy the sample file into the zip
		if _, err := io.Copy(zipEntry, sampleFile); err != nil {
			log.Printf("Failed to copy sample %d to zip: %v", sample.ID, err)
			return fmt.Errorf("failed to copy sample %d to zip: %w", sample.ID, err)
		}

		log.Printf("Successfully added sample %d to zip", sample.ID)
	}

	log.Printf("Successfully created zip file at %s", zipPath)
	return nil
}

func (s *Service) IsUploadAllowedForPack(packID uint) bool {
	if s.cfg.BypassTimeWindows {
		return true
	}

	pack, err := s.GetPack(packID)
	if err != nil {
		return false
	}

	now := time.Now()
	log.Printf("Upload window for pack %d: %s to %s (now: %s)",
		packID,
		pack.UploadStart.Format(time.RFC3339),
		pack.UploadEnd.Format(time.RFC3339),
		now.Format(time.RFC3339))
	return now.After(pack.UploadStart) && now.Before(pack.UploadEnd)
}
