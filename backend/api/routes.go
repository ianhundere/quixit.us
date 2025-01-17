package api

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/models"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine, store storage.Storage, cfg *config.Config) {
	// Initialize routes
	api := r.Group("/api")

	// Sample pack routes
	packs := api.Group("/samples/packs")
	{
		packs.GET("", listPacks)
		packs.GET("/:id", getPack)
		packs.POST("/:id/upload", uploadSample)
		packs.GET("/:id/download", downloadPack)
	}

	// Submission routes
	submissions := api.Group("/submissions")
	{
		submissions.GET("", listSubmissions)
		submissions.POST("", createSubmission)
	}
}

func listPacks(c *gin.Context) {
	// Get user info from context
	_ = c.GetInt("user_id")
	_ = c.GetString("email")

	// Parse dates
	uploadStart, _ := time.Parse("2006-01-02", "2025-01-01")
	uploadEnd, _ := time.Parse("2006-01-02", "2025-01-07")
	startDate, _ := time.Parse("2006-01-02", "2025-01-08")
	endDate, _ := time.Parse("2006-01-02", "2025-01-14")

	pastUploadStart, _ := time.Parse("2006-01-02", "2024-12-01")
	pastUploadEnd, _ := time.Parse("2006-01-02", "2024-12-07")
	pastStartDate, _ := time.Parse("2006-01-02", "2024-12-08")
	pastEndDate, _ := time.Parse("2006-01-02", "2024-12-14")

	// Return mock data for now
	currentPack := models.SamplePack{
		ID:          1,
		Title:       "Current Pack",
		Description: "This is the current sample pack",
		UploadStart: uploadStart,
		UploadEnd:   uploadEnd,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    true,
	}

	pastPack := models.SamplePack{
		ID:          2,
		Title:       "Past Pack 1",
		Description: "This is a past sample pack",
		UploadStart: pastUploadStart,
		UploadEnd:   pastUploadEnd,
		StartDate:   pastStartDate,
		EndDate:     pastEndDate,
		IsActive:    false,
	}

	c.JSON(http.StatusOK, gin.H{
		"currentPack": currentPack,
		"pastPacks":   []models.SamplePack{pastPack},
	})
}

func getPack(c *gin.Context) {
	// Get user info from context
	_ = c.GetInt("user_id")
	_ = c.GetString("email")

	// Parse dates
	uploadStart, _ := time.Parse("2006-01-02", "2025-01-01")
	uploadEnd, _ := time.Parse("2006-01-02", "2025-01-07")
	startDate, _ := time.Parse("2006-01-02", "2025-01-08")
	endDate, _ := time.Parse("2006-01-02", "2025-01-14")

	// Return mock data for now
	pack := models.SamplePack{
		ID:          1,
		Title:       "Current Pack",
		Description: "This is the current sample pack",
		UploadStart: uploadStart,
		UploadEnd:   uploadEnd,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    true,
	}

	c.JSON(http.StatusOK, pack)
}

func uploadSample(c *gin.Context) {
	// Get pack ID from URL
	packID := c.Param("id")
	id, err := strconv.ParseUint(packID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pack ID"})
		return
	}

	// Get the file from the request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	// Get user info from context
	userID := c.GetInt("user_id")
	email := c.GetString("email")

	// Create a temporary file to store the upload
	tmpfile, err := os.CreateTemp("", "sample-*.wav")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temporary file"})
		return
	}
	defer os.Remove(tmpfile.Name())

	// Copy the file data to the temporary file
	if _, err := io.Copy(tmpfile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Return mock sample data
	sample := models.Sample{
		ID:           uint(id),
		Filename:     header.Filename,
		FileURL:      fmt.Sprintf("/api/samples/download/%d", id),
		FileSize:     header.Size,
		UserID:       uint(userID),
		User:         models.User{ID: uint(userID), Email: email},
		SamplePackID: uint(id),
	}

	c.JSON(http.StatusOK, sample)
}

func downloadPack(c *gin.Context) {
	// Get pack ID from URL
	packID := c.Param("id")
	if packID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pack ID is required"})
		return
	}

	// Create a temporary zip file
	tmpFile, err := os.CreateTemp("", "pack-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	defer os.Remove(tmpFile.Name())

	// Create zip writer
	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	// Get all .wav files from storage directory
	files, err := os.ReadDir("storage")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read storage directory"})
		return
	}

	// Add each .wav file to the zip
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".wav" {
			continue
		}

		// Open the file
		f, err := os.Open(filepath.Join("storage", file.Name()))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
			return
		}
		defer f.Close()

		// Create zip entry
		zipFile, err := zipWriter.Create(file.Name())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip entry"})
			return
		}

		// Copy file contents to zip
		if _, err := io.Copy(zipFile, f); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to zip"})
			return
		}
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize zip"})
		return
	}

	// Seek to beginning of file
	if _, err := tmpFile.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare file"})
		return
	}

	// Set headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=pack_%s.zip", packID))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// Stream the file
	c.File(tmpFile.Name())
}

func listSubmissions(c *gin.Context) {
	// Get user info from context
	userID := c.GetInt("user_id")
	email := c.GetString("email")

	// Return mock data for now
	c.JSON(http.StatusOK, []models.Submission{
		{
			ID:          1,
			Title:       "My Submission",
			Description: "This is my submission",
			FileURL:     "http://example.com/submission.zip",
			User: models.User{
				ID:    uint(userID),
				Email: email,
			},
		},
	})
}

func createSubmission(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
