package storage

import (
	"io"
	"os"
	"path/filepath"

	"sample-exchange/backend/config"
)

type Storage interface {
	SaveSample(file io.Reader, filename string) (string, error)
	SaveSubmission(file io.Reader, filename string) (string, error)
	Delete(filepath string) error
}

type FileStorage struct {
	samplePath     string
	submissionPath string
}

func NewStorage(cfg *config.Config) Storage {
	return &FileStorage{
		samplePath:     filepath.Join(cfg.StoragePath, "samples"),
		submissionPath: filepath.Join(cfg.StoragePath, "submissions"),
	}
}

func (s *FileStorage) SaveSample(file io.Reader, filename string) (string, error) {
	return s.saveFile(s.samplePath, file, filename)
}

func (s *FileStorage) SaveSubmission(file io.Reader, filename string) (string, error) {
	return s.saveFile(s.submissionPath, file, filename)
}

func (s *FileStorage) Delete(filepath string) error {
	return os.Remove(filepath)
}

func (s *FileStorage) saveFile(basePath string, file io.Reader, filename string) (string, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", err
	}

	// Create file path
	filePath := filepath.Join(basePath, filename)

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", err
	}

	return filePath, nil
}
