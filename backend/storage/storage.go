package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sample-exchange/backend/config"

	"github.com/google/uuid"
)

type Storage struct {
	basePath string
}

func NewStorage(cfg *config.Config) *Storage {
	return &Storage{
		basePath: cfg.StoragePath,
	}
}

func (s *Storage) SaveSample(file io.Reader, filename string) (string, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Generate unique filename to avoid collisions
	uniqueFilename := fmt.Sprintf("%s_%s", uuid.New().String(), filename)
	filePath := filepath.Join(s.basePath, uniqueFilename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the file data
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

func (s *Storage) GetSample(filepath string) (io.ReadCloser, error) {
	return os.Open(filepath)
}

func (s *Storage) DeleteSample(filepath string) error {
	return os.Remove(filepath)
} 