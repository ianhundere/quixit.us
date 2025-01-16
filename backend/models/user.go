package models

import (
	"fmt"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"`
	Verified     bool       `json:"-" gorm:"default:false"`
	VerifyToken  string     `json:"-" gorm:"size:255"`
	RefreshToken string     `json:"-" gorm:"size:500"`
	LoginAttempts int      `json:"-" gorm:"default:0"`
	LockedUntil   *time.Time `json:"-"`
}

// PasswordRequirements defines the requirements for password complexity
var PasswordRequirements = struct {
	MinLength  int
	MinUpper   int
	MinLower   int
	MinNumber  int
	MinSpecial int
}{
	MinLength:  8,
	MinUpper:   1,
	MinLower:   1,
	MinNumber:  1,
	MinSpecial: 1,
}

func ValidatePassword(password string) error {
	var (
		upper   int
		lower   int
		number  int
		special int
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			upper++
		case unicode.IsLower(char):
			lower++
		case unicode.IsNumber(char):
			number++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			special++
		}
	}

	if len(password) < PasswordRequirements.MinLength {
		return fmt.Errorf("password must be at least %d characters long", PasswordRequirements.MinLength)
	}
	if upper < PasswordRequirements.MinUpper {
		return fmt.Errorf("password must contain at least %d uppercase letter", PasswordRequirements.MinUpper)
	}
	if lower < PasswordRequirements.MinLower {
		return fmt.Errorf("password must contain at least %d lowercase letter", PasswordRequirements.MinLower)
	}
	if number < PasswordRequirements.MinNumber {
		return fmt.Errorf("password must contain at least %d number", PasswordRequirements.MinNumber)
	}
	if special < PasswordRequirements.MinSpecial {
		return fmt.Errorf("password must contain at least %d special character", PasswordRequirements.MinSpecial)
	}

	return nil
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
} 