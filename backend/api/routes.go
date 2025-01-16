package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sample-exchange/backend/auth"
	"sample-exchange/backend/config"
	"sample-exchange/backend/db"
	"sample-exchange/backend/email"
	"sample-exchange/backend/middleware"
	"sample-exchange/backend/models"
	"sample-exchange/backend/services/samplepack"
	"sample-exchange/backend/services/submission"
	"sample-exchange/backend/storage"

	"sample-exchange/backend/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var store *storage.Storage
var emailService *email.EmailService
var rateLimiter *middleware.RateLimiter
var packService *samplepack.Service
var submissionService *submission.Service

func Init(r *gin.Engine, s *storage.Storage, cfg *config.Config) {
	store = s
	emailService = email.NewEmailService(cfg)
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
	
	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", registerUser)
		auth.POST("/login", loginUser)
		auth.POST("/refresh", refreshToken)
		auth.GET("/verify/:token", verifyEmail)
	}

	// Sample routes
	samples := api.Group("/samples")
	samples.Use(authMiddleware()) // Protect these routes
	{
		samples.GET("/packs", listSamplePacks)
		samples.GET("/packs/:id", getSamplePack)
		samples.POST("/upload", middleware.ValidateFileUpload(), uploadSample)
		samples.GET("/download/:id", downloadSample)
	}

	// Submission routes
	submissions := api.Group("/submissions")
	submissions.Use(authMiddleware())
	{
		submissions.POST("/", createSubmission)
		submissions.GET("/", listSubmissions)
		submissions.GET("/:id", getSubmission)
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
	var register struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password complexity
	if err := models.ValidatePassword(register.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.DB.Where("email = ?", register.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Generate verification token
	verifyToken := uuid.New().String()

	user := models.User{
		Email: register.Email,
		VerifyToken: verifyToken,
	}

	if err := user.SetPassword(register.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Send verification email
	if err := emailService.SendVerificationEmail(user.Email, verifyToken); err != nil {
		log.Printf("Warning: Failed to send verification email: %v", err)
		// For development, auto-verify the user
		user.Verified = true
		user.VerifyToken = ""
		if err := db.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
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
	if !packService.IsUploadAllowed() {
		c.Error(errors.NewAuthorizationError("Upload window is closed"))
		return
	}

	currentPack, err := packService.GetCurrentPack()
	if err != nil {
		c.Error(errors.NewNotFoundError("Active sample pack"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.Error(errors.NewValidationError("file", "No file provided"))
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer src.Close()

	// Save the file using our storage package
	filePath, err := store.SaveSample(src, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	sample := models.Sample{
		Filename:   file.Filename,
		FileSize:   file.Size,
		UploadedAt: time.Now(),
		FilePath:   filePath,
		UserID:     c.GetUint("userID"),
		SamplePackID: currentPack.ID,
	}

	// Save to database
	if err := db.DB.Create(&sample).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to database"})
		return
	}

	if err := packService.AddSample(currentPack.ID, &sample); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add sample to pack"})
		return
	}

	c.JSON(http.StatusCreated, sample)
}

func downloadSample(c *gin.Context) {
	// TODO: Implement file download logic
	c.JSON(http.StatusOK, gin.H{"message": "Download endpoint"})
}

func createSubmission(c *gin.Context) {
	var submission models.Submission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	err := submissionService.CreateSubmission(userID, &submission)

	if err != nil {
		switch err {
		case models.ErrSubmissionClosed:
			c.JSON(http.StatusForbidden, gin.H{"error": "Submission window is closed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
		}
		return
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
	currentPack, _ := packService.GetCurrentPack()
	pastPacks, err := packService.ListPacks(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sample packs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currentPack": currentPack,
		"pastPacks":   pastPacks,
		"canUpload":   packService.IsUploadAllowed(),
		"canSubmit":   packService.IsSubmissionAllowed(),
	})
}

func getSamplePack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pack ID"})
		return
	}

	pack, err := packService.GetPack(uint(id))
	if err != nil {
		switch err {
		case models.ErrPackNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Sample pack not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sample pack"})
		}
		return
	}
	
	c.JSON(http.StatusOK, pack)
}

func verifyEmail(c *gin.Context) {
	token := c.Param("token")
	
	var user models.User
	if err := db.DB.Where("verify_token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid verification token"})
		return
	}

	user.Verified = true
	user.VerifyToken = ""
	
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
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

func voteSubmission(c *gin.Context) {
	var vote struct {
		Value int `json:"value" binding:"required,oneof=-1 1"`
	}

	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submissionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	userID := c.GetUint("userID")
	err = submissionService.Vote(userID, uint(submissionID), vote.Value)

	if err != nil {
		switch err {
		case models.ErrAlreadyVoted:
			c.JSON(http.StatusConflict, gin.H{"error": "Already voted for this submission"})
		case submission.ErrInvalidVoteValue:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote value"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
} 