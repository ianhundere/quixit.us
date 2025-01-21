package testdata

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sample-exchange/backend/auth"
	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/services/samplepack"
	"sample-exchange/backend/services/submission"
	"sample-exchange/backend/services/user"
)

type TestData struct {
	Token        string
	UserID       uint
	PackID       uint
	SubmissionID uint
}

// Setup creates all necessary test data and returns test info
func Setup(cfg *config.Config) (*TestData, error) {
	// Initialize DB if needed
	if err := db.SetupDB(); err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	// Ensure storage directories exist
	if err := setupStorage(); err != nil {
		return nil, fmt.Errorf("failed to setup storage: %w", err)
	}

	// Set JWT secret from config
	auth.SetJWTSecret(cfg.JWTSecret)

	// Create services
	userSvc := user.NewService()
	packSvc := samplepack.NewService(cfg)
	submissionSvc := submission.NewService(cfg, packSvc)

	// Create test user
	testUser, err := userSvc.CreateTestUser()
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	// Create test pack
	pack, err := packSvc.CreateTestPack(testUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create test pack: %w", err)
	}

	// Create test submission
	submission, err := submissionSvc.CreateTestSubmission(testUser.ID, pack.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create test submission: %w", err)
	}

	// Generate JWT token for test user
	token, err := auth.GenerateToken(testUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	testData := &TestData{
		Token:        token,
		UserID:       testUser.ID,
		PackID:       pack.ID,
		SubmissionID: submission.ID,
	}

	log.Printf("Created test data: user=%d, pack=%d, submission=%d", 
		testData.UserID, testData.PackID, testData.SubmissionID)

	// Update Bruno environment
	if err := testData.UpdateBrunoEnv(); err != nil {
		log.Printf("Warning: failed to update Bruno env: %v", err)
	}

	return testData, nil
}

func setupStorage() error {
	// Create storage directories
	dirs := []string{
		"storage/samples",
		"storage/submissions",
		"/tmp",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create test files directly in /tmp
	files := map[string]string{
		"test_sample.wav": "/tmp/test_sample.wav",
		"test_submission.mp3": "/tmp/test_submission.mp3",
	}
	for name, dst := range files {
		if err := createTestFile(name, dst); err != nil {
			return fmt.Errorf("failed to create %s: %w", dst, err)
		}
	}

	return nil
}

// createTestFile creates a minimal valid audio file
func createTestFile(name, dst string) error {
	var content []byte
	if filepath.Ext(name) == ".wav" {
		// Create minimal valid WAV file (44.1kHz, 16-bit, mono)
		content = []byte{
			// RIFF header
			0x52, 0x49, 0x46, 0x46, // "RIFF"
			0x24, 0x00, 0x00, 0x00, // Chunk size (36 bytes)
			0x57, 0x41, 0x56, 0x45, // "WAVE"
			// fmt chunk
			0x66, 0x6d, 0x74, 0x20, // "fmt "
			0x10, 0x00, 0x00, 0x00, // Chunk size (16 bytes)
			0x01, 0x00,             // Audio format (1 = PCM)
			0x01, 0x00,             // Number of channels (1)
			0x44, 0xac, 0x00, 0x00, // Sample rate (44100)
			0x88, 0x58, 0x01, 0x00, // Byte rate (44100 * 2)
			0x02, 0x00,             // Block align
			0x10, 0x00,             // Bits per sample (16)
			// data chunk
			0x64, 0x61, 0x74, 0x61, // "data"
			0x00, 0x00, 0x00, 0x00, // Chunk size (0 bytes)
		}
	} else if filepath.Ext(name) == ".mp3" {
		// Create minimal valid MP3 file (empty MPEG frame)
		content = []byte{
			// ID3v2 header
			0x49, 0x44, 0x33,       // "ID3"
			0x03, 0x00,             // Version 2.3.0
			0x00,                   // No flags
			0x00, 0x00, 0x00, 0x00, // Size (0 bytes)
			// MPEG frame header
			0xFF, 0xFB, 0x90, 0x64, // MPEG-1 Layer 3, 44.1kHz
			0x00,                   // Empty frame
		}
	}

	return os.WriteFile(dst, content, 0644)
}

func (td *TestData) UpdateBrunoEnv() error {
	// Read Bruno env file
	envPath := "bruno-collection/environments/Dev.bru"
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read Bruno env file: %w", err)
	}

	// Replace token
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "auth_token:") {
			lines[i] = fmt.Sprintf("  auth_token: %s", td.Token)
		}
	}

	// Write back
	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
}
