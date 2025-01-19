package db

import (
	"fmt"

	"sample-exchange/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetupDB() error {
	var err error
	dsn := "host=localhost user=postgres password=postgres dbname=sample_exchange port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.SamplePack{},
		&models.Sample{},
		&models.Submission{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create indexes
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}

func createIndexes(db *gorm.DB) error {
	// Add indexes for User
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		return err
	}

	// Add indexes for SamplePack
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_sample_packs_is_active ON sample_packs(is_active)").Error; err != nil {
		return err
	}

	// Add indexes for Sample
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_samples_user_id ON samples(user_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_samples_sample_pack_id ON samples(sample_pack_id)").Error; err != nil {
		return err
	}

	// Add indexes for Submission
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON submissions(user_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_submissions_sample_pack_id ON submissions(sample_pack_id)").Error; err != nil {
		return err
	}

	return nil
}
