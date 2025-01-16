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
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrSubmissionClosed  = errors.New("submission window is closed")
	ErrInvalidVoteValue  = errors.New("invalid vote value")
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
		Preload("Votes").
		First(&submission, id).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrSubmissionNotFound
	}
	return &submission, err
}

func (s *Service) ListSubmissions(packID uint, limit, offset int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.db.Where("sample_pack_id = ?", packID).
		Preload("User").
		Preload("Votes").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&submissions).Error

	// Calculate vote counts
	for i := range submissions {
		var count int64
		s.db.Model(&models.Vote{}).
			Where("submission_id = ?", submissions[i].ID).
			Select("COALESCE(SUM(value), 0)").
			Scan(&count)
		submissions[i].VoteCount = int(count)
	}

	return submissions, err
}

func (s *Service) Vote(userID, submissionID uint, value int) error {
	if value != -1 && value != 1 {
		return ErrInvalidVoteValue
	}

	// Check if user has already voted
	var existingVote models.Vote
	err := s.db.Where("user_id = ? AND submission_id = ?", userID, submissionID).
		First(&existingVote).Error

	if err == nil {
		// Update existing vote
		if existingVote.Value == value {
			return models.ErrAlreadyVoted
		}
		existingVote.Value = value
		return s.db.Save(&existingVote).Error
	}

	// Create new vote
	vote := models.Vote{
		UserID:       userID,
		SubmissionID: submissionID,
		Value:        value,
	}
	return s.db.Create(&vote).Error
}

func (s *Service) GetTopSubmissions(packID uint, limit int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.db.Where("sample_pack_id = ?", packID).
		Preload("User").
		Preload("Votes").
		Joins("LEFT JOIN votes ON votes.submission_id = submissions.id").
		Group("submissions.id").
		Order("SUM(COALESCE(votes.value, 0)) DESC").
		Limit(limit).
		Find(&submissions).Error

	// Calculate vote counts
	for i := range submissions {
		var count int64
		s.db.Model(&models.Vote{}).
			Where("submission_id = ?", submissions[i].ID).
			Select("COALESCE(SUM(value), 0)").
			Scan(&count)
		submissions[i].VoteCount = int(count)
	}

	return submissions, err
} 