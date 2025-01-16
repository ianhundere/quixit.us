package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sample-exchange/backend/auth"
	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/middleware"
	"sample-exchange/backend/models"
	"sample-exchange/backend/services/samplepack"
	"sample-exchange/backend/services/submission"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var store *storage.Storage
var rateLimiter *middleware.RateLimiter
var packService *samplepack.Service
var submissionService *submission.Service

const maxFileSize = 50 * 1024 * 1024 // 50MB

func Init(r *gin.Engine, s *storage.Storage, cfg *config.Config) {
	store = s
	rateLimiter = middleware.NewRateLimiter(time.Minute, 60) // 60 requests per minute
	packService = samplepack.NewService(cfg)
	submissionService = submission.NewService(cfg, packService)
	
	auth.Init(cfg)
	r.Use(middleware.ErrorHandler())
	SetupRoutes(r)
}

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(rateLimiter.RateLimit())
	
	// Health check endpoint for monitoring
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", registerUser)
		auth.POST("/login", loginUser)
		auth.POST("/refresh", refreshToken)
	}

	// Sample routes
	samples := api.Group("/samples")
	samples.Use(authMiddleware()) // Protect these routes
	{
		samples.GET("/packs", listSamplePacks)
		samples.GET("/packs/:id", getSamplePack)
		samples.POST("/packs/:id/upload", middleware.ValidateFileUpload(), uploadSample)
		samples.GET("/download/:id", downloadSample)
	}

	// Admin routes for managing packs
	admin := api.Group("/admin")
	admin.Use(authMiddleware())
	{
		admin.POST("/packs", createSamplePack)
		admin.PATCH("/packs/:id/windows", updatePackWindows)
	}

	// Submission routes
	submissions := api.Group("/submissions")
	submissions.Use(authMiddleware())
	{
		submissions.POST("", middleware.ValidateFileUpload(), createSubmission)
		submissions.GET("", listSubmissions)
		submissions.GET("/:id", getSubmission)
		submissions.GET("/:id/download", downloadSubmission)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		claims, err := auth.ValidateToken(tokenParts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Add user info to context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		
		c.Next()
	}
}

func registerUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Create user
	user := &models.User{
		Email: input.Email,
	}

	if err := user.SetPassword(input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Save to database
	if err := db.DB.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

func loginUser(c *gin.Context) {
	var login struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", login.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(login.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update refresh token in database
	user.RefreshToken = refreshToken
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func uploadSample(c *gin.Context) {
	// Get pack ID from URL
	packID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	// Get pack
	pack, err := packService.GetPack(uint(packID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pack"})
		return
	}

	if !pack.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pack is not active"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	// Open and read file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer src.Close()

	// Save file to storage
	filePath, err := store.SaveSample(src, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create sample record
	sample := &models.Sample{
		SamplePackID: uint(packID),
		UserID:      c.GetUint("userID"),
		Filename:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
		UploadedAt:  time.Now(),
	}

	if err := db.DB.Create(sample).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to database"})
		return
	}

	// Load relations
	if err := db.DB.Preload("User").First(sample, sample.ID).Error; err != nil {
		c.JSON(http.StatusOK, sample)
		return
	}

	sample.FileURL = fmt.Sprintf("/api/samples/download/%d", sample.ID)
	c.JSON(http.StatusOK, sample)
}

func downloadSample(c *gin.Context) {
	// First try auth header
	userID := c.GetUint("userID")
	
	// If no user ID from auth header, try token from query params
	if userID == 0 {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication provided"})
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			log.Printf("Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		userID = claims.UserID
		log.Printf("Authenticated with token for user: %d", userID)
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sample ID"})
		return
	}

	var sample models.Sample
	if err := db.DB.First(&sample, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sample not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(sample.FilePath); os.IsNotExist(err) {
		log.Printf("File not found at path: %s", sample.FilePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Set content type based on file extension
	extension := strings.ToLower(filepath.Ext(sample.Filename))
	switch extension {
	case ".wav":
		c.Header("Content-Type", "audio/wav")
	case ".mp3":
		c.Header("Content-Type", "audio/mpeg")
	case ".aiff":
		c.Header("Content-Type", "audio/aiff")
	case ".flac":
		c.Header("Content-Type", "audio/flac")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}

	// Set headers for streaming
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, sample.Filename))

	// Serve the file
	c.File(sample.FilePath)
}

func createSubmission(c *gin.Context) {
	// Get form data
	title := c.PostForm("title")
	description := c.PostForm("description")
	samplePackID, err := strconv.ParseUint(c.PostForm("samplePackId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer src.Close()

	// Save the file
	filePath, err := store.SaveSample(src, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	submission := models.Submission{
		Title:       title,
		Description: description,
		FilePath:    filePath,
		FileSize:    file.Size,
		UserID:      c.GetUint("userID"),
		SamplePackID: uint(samplePackID),
		SubmittedAt: time.Now(),
	}

	userID := c.GetUint("userID")
	err = submissionService.CreateSubmission(userID, &submission)
	if err != nil {
		switch err {
		case models.ErrSubmissionClosed:
			c.JSON(http.StatusForbidden, gin.H{"error": "Submission window is closed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
		}
		return
	}

	// Generate file URL
	submission.FileURL = fmt.Sprintf("/api/submissions/%d/download", submission.ID)

	// Load relations
	if err := db.DB.Preload("User").Preload("SamplePack").First(&submission, submission.ID).Error; err != nil {
		log.Printf("Warning: Failed to load relations: %v", err)
	}

	c.JSON(http.StatusCreated, submission)
}

func listSubmissions(c *gin.Context) {
	packID, err := strconv.ParseUint(c.Query("pack_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	submissions, err := submissionService.ListSubmissions(uint(packID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

func getSubmission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	submission, err := submissionService.GetSubmission(uint(id))
	if err != nil {
		switch err {
		case models.ErrSubmissionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submission"})
		}
		return
	}

	c.JSON(http.StatusOK, submission)
}

func listSamplePacks(c *gin.Context) {
	// Get current pack
	currentPack, err := packService.GetCurrentPack()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current pack"})
		return
	}

	// Get past packs
	pastPacks, err := packService.ListPacks(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list packs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentPack": currentPack,
		"pastPacks":   pastPacks,
	})
}

func getSamplePack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	// Get pack with samples preloaded
	var pack models.SamplePack
	err = db.DB.Preload("Samples").Preload("Samples.User").First(&pack, id).Error
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sample pack not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sample pack"})
		return
	}

	// Generate file URLs for samples
	for i := range pack.Samples {
		pack.Samples[i].FileURL = fmt.Sprintf("/api/samples/download/%d", pack.Samples[i].ID)
	}
	
	c.JSON(http.StatusOK, pack)
}

func refreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if claims.Type != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update refresh token in database
	user.RefreshToken = refreshToken
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func downloadSubmission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var submission models.Submission
	if err := db.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(submission.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.FileAttachment(submission.FilePath, fmt.Sprintf("submission_%d.wav", submission.ID))
}

func createSamplePack(c *gin.Context) {
	var pack struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&pack); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Creating new pack with title: %s", pack.Title)
	newPack, err := packService.CreatePack()
	if err != nil {
		log.Printf("Failed to create pack: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sample pack"})
		return
	}

	newPack.Title = pack.Title
	newPack.Description = pack.Description

	if err := db.DB.Save(newPack).Error; err != nil {
		log.Printf("Failed to save pack: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save sample pack"})
		return
	}

	log.Printf("Successfully created pack: %+v", newPack)
	c.JSON(http.StatusCreated, newPack)
}

func updatePackWindows(c *gin.Context) {
	var req struct {
		UploadStart *time.Time `json:"uploadStart"`
		UploadEnd   *time.Time `json:"uploadEnd"`
		StartDate   *time.Time `json:"startDate"`
		EndDate     *time.Time `json:"endDate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	var pack models.SamplePack
	if err := db.DB.First(&pack, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pack not found"})
		return
	}

	// Update only provided fields
	if req.UploadStart != nil {
		pack.UploadStart = *req.UploadStart
	}
	if req.UploadEnd != nil {
		pack.UploadEnd = *req.UploadEnd
	}
	if req.StartDate != nil {
		pack.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		pack.EndDate = *req.EndDate
	}

	if err := db.DB.Save(&pack).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pack"})
		return
	}

	c.JSON(http.StatusOK, pack)
} 