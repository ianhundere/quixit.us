package middleware

import (
	"errors"
	"log"
	"net/http"

	customerrors "sample-exchange/backend/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorHandler is a middleware that handles all errors in a consistent way
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only handle errors if we have any
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		handleError(c, err)
	}
}

func handleError(c *gin.Context, err error) {
	var apiErr *customerrors.APIError
	if errors.As(err, &apiErr) {
		// Handle our custom API errors
		if apiErr.Internal != nil {
			log.Printf("Internal error: %v", apiErr.Internal)
		}
		c.JSON(apiErr.Code, apiErr)
		return
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		// Handle validation errors
		details := make([]map[string]string, 0)
		for _, err := range validationErrors {
			details = append(details, map[string]string{
				"field":   err.Field(),
				"message": formatValidationError(err),
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"type":    customerrors.TypeValidation,
			"message": "Validation failed",
			"details": details,
		})
		return
	}

	// Handle unknown errors
	log.Printf("Unhandled error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"type":    customerrors.TypeInternal,
		"message": "An internal error occurred",
	})
}

func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	default:
		return "Invalid value"
	}
}
