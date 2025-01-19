package user

import (
	"errors"

	"sample-exchange/backend/db"
	"sample-exchange/backend/models"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService() *Service {
	return &Service{
		db: db.GetDB(),
	}
}

// GetByID retrieves a user by their ID
func (s *Service) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email
func (s *Service) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetOrCreateOAuthUser gets an existing user by provider and email or creates a new one
func (s *Service) GetOrCreateOAuthUser(email, name, provider, avatar string) (*models.User, error) {
	var user models.User

	// Try to find existing user by email
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new user
		user = models.User{
			Email:    email,
			Name:     name,
			Provider: provider,
			Avatar:   avatar,
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}
	} else {
		// Update existing user's OAuth info
		user.Name = name
		user.Provider = provider
		user.Avatar = avatar
		if err := s.db.Save(&user).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// List retrieves a paginated list of users
func (s *Service) List(limit, offset int) ([]models.User, error) {
	var users []models.User
	if err := s.db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Delete removes a user by their ID
func (s *Service) Delete(id uint) error {
	return s.db.Delete(&models.User{}, id).Error
}
