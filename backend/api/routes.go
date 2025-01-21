package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/middleware"
	"sample-exchange/backend/models"
	"sample-exchange/backend/services/samplepack"
	"sample-exchange/backend/services/submission"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	packService       *samplepack.Service
	submissionService *submission.Service
	storage           storage.Storage
	config            *config.Config
}

func NewHandler(packService *samplepack.Service, submissionService *submission.Service, storage storage.Storage, cfg *config.Config) *Handler {
	return &Handler{
		packService:       packService,
		submissionService: submissionService,
		storage:           storage,
		config:            cfg,
	}
}

func Init(r *gin.Engine, store storage.Storage, cfg *config.Config) {
	packService := samplepack.NewService(cfg)
	submissionService := submission.NewService(cfg, packService)
	handler := NewHandler(packService, submissionService, store, cfg)

	// Initialize routes
	api := r.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.GET("/current-user", middleware.Auth(), func(c *gin.Context) {
			userID := uint(c.GetInt("user_id"))
			email := c.GetString("email")
			c.JSON(http.StatusOK, gin.H{
				"ID":    userID,
				"email": email,
			})
		})
	}

	// Admin routes for pack management
	admin := api.Group("/admin")
	{
		admin.POST("/packs", middleware.Auth(), middleware.RequireAdmin(), handler.createNewPack)
		admin.POST("/packs/:id/close", middleware.Auth(), middleware.RequireAdmin(), handler.closePack)
	}

	// Sample pack routes
	packs := api.Group("/samples/packs")
	{
		packs.GET("", handler.listPacks)
		packs.GET("/:id", handler.getPack)
		packs.POST("/:id/upload", middleware.Auth(), middleware.ValidateFileUpload(), handler.uploadSample)
		packs.GET("/:id/download", handler.downloadPack)
	}

	// Submission routes
	submissions := api.Group("/submissions")
	{
		submissions.GET("", middleware.Auth(), handler.listSubmissions)
		submissions.POST("", middleware.Auth(), middleware.ValidateFileUpload(), handler.createSubmission)
		submissions.GET("/:id", middleware.Auth(), handler.getSubmission)
		submissions.GET("/:id/download", middleware.Auth(), handler.downloadSubmission)
	}
}

func (h *Handler) listPacks(c *gin.Context) {
	currentPack, err := h.packService.GetCurrentPack()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current pack"})
		return
	}

	pastPacks, err := h.packService.ListPacks(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list past packs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentPack": currentPack,
		"pastPacks":   pastPacks,
	})
}

func (h *Handler) getPack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	pack, err := h.packService.GetPack(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	c.JSON(http.StatusOK, pack)
}

func (h *Handler) uploadSample(c *gin.Context) {
	packID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	if !h.packService.IsUploadAllowedForPack(uint(packID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Upload window is closed"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	userID := uint(c.GetInt("user_id"))

	// Store the file using the storage interface
	filePath, err := h.storage.SaveSample(file, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}

	sample := &models.Sample{
		Filename:     header.Filename,
		FilePath:     filePath,
		FileSize:     header.Size,
		UserID:       userID,
		SamplePackID: uint(packID),
	}

	if err := h.packService.AddSample(uint(packID), sample); err != nil {
		h.storage.Delete(filePath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add sample"})
		return
	}

	sample.FileURL = fmt.Sprintf("/api/samples/packs/%d/samples/%d/download", packID, sample.ID)
	c.JSON(http.StatusOK, sample)
}

func (h *Handler) downloadPack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	pack, err := h.packService.GetPack(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	// Create temporary zip file
	zipPath := fmt.Sprintf("/tmp/pack_%d.zip", id)
	if err := h.packService.CreatePackZip(*pack, zipPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip file"})
		return
	}
	defer h.storage.Delete(zipPath)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=pack_%d.zip", id))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.File(zipPath)
}

func (h *Handler) listSubmissions(c *gin.Context) {
	packID, err := strconv.ParseUint(c.Query("pack_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	limit := 10
	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	submissions, err := h.submissionService.ListSubmissions(uint(packID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list submissions"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

func (h *Handler) createSubmission(c *gin.Context) {
	// Get file from form data
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Get other form fields
	packID, err := strconv.ParseUint(c.Request.FormValue("sample_pack_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	userID := uint(c.GetInt("user_id"))

	// Store the file using the storage interface
	filePath, err := h.storage.SaveSubmission(file, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file"})
		return
	}

	// Create submission record
	submission := &models.Submission{
		Title:        c.Request.FormValue("title"),
		Filename:     header.Filename,
		FilePath:     filePath,
		FileSize:     header.Size,
		UserID:       userID,
		SamplePackID: uint(packID),
		SubmittedAt:  time.Now(),
	}

	if err := h.submissionService.CreateSubmission(userID, submission); err != nil {
		h.storage.Delete(filePath) // Clean up on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	submission.FileURL = fmt.Sprintf("/api/submissions/%d/download", submission.ID)
	c.JSON(http.StatusCreated, submission)
}

func (h *Handler) getSubmission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	submission, err := h.submissionService.GetSubmission(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, submission)
}

func (h *Handler) downloadSubmission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	submission, err := h.submissionService.GetSubmission(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", submission.Filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.File(submission.FilePath)
}

func (h *Handler) createNewPack(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	pack, err := h.packService.CreatePack()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pack"})
		return
	}

	pack.Title = req.Title
	pack.Description = req.Description

	if err := db.GetDB().Save(pack).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save pack"})
		return
	}

	c.JSON(http.StatusCreated, pack)
}

func (h *Handler) closePack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	pack, err := h.packService.GetPack(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	pack.IsActive = false
	if err := db.GetDB().Save(pack).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close pack"})
		return
	}

	c.JSON(http.StatusOK, pack)
}
