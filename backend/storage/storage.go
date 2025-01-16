package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"sample-exchange/backend/config"
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
	// Create year/month based directory structure
	now := time.Now()
	dirPath := filepath.Join(s.basePath, "samples", fmt.Sprintf("%d/%02d", now.Year(), now.Month()))
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	filePath := filepath.Join(dirPath, fmt.Sprintf("%d_%s", now.Unix(), filename))

	// Create the file
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