package api

import (
	"net/http"
	"strings"
	"time"

	"sample-exchange/backend/auth"
	"sample-exchange/backend/db"
	"sample-exchange/backend/models"
	"sample-exchange/backend/storage"

	"github.com/gin-gonic/gin"
)

var store *storage.Storage

func InitRoutes(r *gin.Engine, s *storage.Storage) {
	store = s
	SetupRoutes(r)
}

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	
	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", registerUser)
		auth.POST("/login", loginUser)
	}

	// Sample routes
	samples := api.Group("/samples")
	samples.Use(authMiddleware()) // Protect these routes
	{
		samples.GET("/packs", listSamplePacks)
		samples.GET("/packs/:id", getSamplePack)
		samples.POST("/upload", uploadSample)
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

	// Check if user already exists
	var existingUser models.User
	if err := db.DB.Where("email = ?", register.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	user := models.User{
		Email: register.Email,
	}

	if err := user.SetPassword(register.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := auth.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
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

	token, err := auth.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func uploadSample(c *gin.Context) {
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

	// Save the file using our storage package
	filePath, err := store.SaveSample(src, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	sample := models.Sample{
		Filename: file.Filename,
		FileSize: file.Size,
		UploadedAt: time.Now(),
		FilePath: filePath,
	}

	// Save to database
	if err := db.DB.Create(&sample).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to database"})
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
	submission.SubmittedAt = time.Now()
	// TODO: Save submission
	c.JSON(http.StatusCreated, submission)
}

func listSubmissions(c *gin.Context) {
	// TODO: Implement submission listing
	c.JSON(http.StatusOK, []models.Submission{})
}

func getSubmission(c *gin.Context) {
	id := c.Param("id")
	// TODO: Fetch submission by ID
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// SamplePackResponse represents the API response for sample packs
type SamplePackResponse struct {
	CurrentPack *models.SamplePack   `json:"currentPack,omitempty"`
	PastPacks  []models.SamplePack   `json:"pastPacks"`
}

func listSamplePacks(c *gin.Context) {
	// TODO: Implement database query
	response := SamplePackResponse{
		CurrentPack: &models.SamplePack{
			StartDate: time.Now(),
			EndDate:   time.Now().Add(72 * time.Hour),
			IsActive:  true,
			Samples:   []models.Sample{},
		},
		PastPacks: []models.SamplePack{},
	}
	
	c.JSON(http.StatusOK, response)
}

func getSamplePack(c *gin.Context) {
	id := c.Param("id")
	// TODO: Implement database query
	pack := models.SamplePack{
		StartDate: time.Now(),
		EndDate:   time.Now().Add(72 * time.Hour),
		IsActive:  true,
		Samples:   []models.Sample{},
	}
	
	c.JSON(http.StatusOK, pack)
} 