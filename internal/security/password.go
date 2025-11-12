// internal/security/password.go - Enhanced password security
package security

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 12
	MaxPasswordLength = 128
	BcryptCost        = 12 // Increased from default 10
)

var (
	ErrPasswordTooShort      = errors.New("password must be at least 12 characters long")
	ErrPasswordTooLong       = errors.New("password must be less than 128 characters")
	ErrPasswordNoUppercase   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial     = errors.New("password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	ErrPasswordCommon        = errors.New("password is too common and easily guessable")
)

// CommonPasswords - list of commonly used passwords to reject
var CommonPasswords = map[string]bool{
	"password123456": true,
	"123456789012": true,
	"qwerty123456": true,
	"admin123456": true,
	"welcome12345": true,
	"password1234": true,
	"letmein12345": true,
	// Add more common passwords
}

// PasswordRequirements holds password validation rules
type PasswordRequirements struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
	ForbidCommon     bool
}

// DefaultPasswordRequirements returns production-grade password rules
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:        MinPasswordLength,
		MaxLength:        MaxPasswordLength,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
		ForbidCommon:     true,
	}
}

// ValidatePassword checks if password meets all security requirements
func ValidatePassword(password string, requirements PasswordRequirements) error {
	// Check length
	if len(password) < requirements.MinLength {
		return ErrPasswordTooShort
	}
	if len(password) > requirements.MaxLength {
		return ErrPasswordTooLong
	}

	// Check for common passwords
	if requirements.ForbidCommon && CommonPasswords[password] {
		return ErrPasswordCommon
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	// Check character requirements
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if requirements.RequireUppercase && !hasUpper {
		return ErrPasswordNoUppercase
	}
	if requirements.RequireLowercase && !hasLower {
		return ErrPasswordNoLowercase
	}
	if requirements.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}
	if requirements.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// HashPassword securely hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	// Validate before hashing
	if err := ValidatePassword(password, DefaultPasswordRequirements()); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// ComparePassword compares a password with its hash
func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// PasswordStrength returns a score from 0-100 indicating password strength
func PasswordStrength(password string) int {
	score := 0

	// Length score (up to 30 points)
	if len(password) >= 12 {
		score += 10
	}
	if len(password) >= 16 {
		score += 10
	}
	if len(password) >= 20 {
		score += 10
	}

	// Character variety (up to 40 points)
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	if hasUpper {
		score += 10
	}
	if hasLower {
		score += 10
	}
	if hasDigit {
		score += 10
	}
	if hasSpecial {
		score += 10
	}

	// Complexity patterns (up to 30 points)
	// No repeated characters
	if !regexp.MustCompile(`(.)\1{2,}`).MatchString(password) {
		score += 10
	}
	// Mix of character types in sequence
	if regexp.MustCompile(`[a-z][A-Z]|[A-Z][a-z]`).MatchString(password) {
		score += 10
	}
	// Has special characters not at start/end
	if regexp.MustCompile(`^[a-zA-Z0-9].*[^a-zA-Z0-9].*[a-zA-Z0-9]$`).MatchString(password) {
		score += 10
	}

	// Penalize common patterns
	if CommonPasswords[password] {
		score = 0
	}
	if regexp.MustCompile(`^[0-9]+$`).MatchString(password) {
		score /= 2
	}
	if regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(password) {
		score -= 10
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// SuggestPasswordImprovement provides feedback for password improvement
func SuggestPasswordImprovement(password string) []string {
	var suggestions []string

	if len(password) < MinPasswordLength {
		suggestions = append(suggestions, 
			"Increase length to at least 12 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		suggestions = append(suggestions, "Add uppercase letters")
	}
	if !hasLower {
		suggestions = append(suggestions, "Add lowercase letters")
	}
	if !hasDigit {
		suggestions = append(suggestions, "Add numbers")
	}
	if !hasSpecial {
		suggestions = append(suggestions, "Add special characters (!@#$%^&*)")
	}

	if regexp.MustCompile(`(.)\1{2,}`).MatchString(password) {
		suggestions = append(suggestions, "Avoid repeating characters")
	}

	if CommonPasswords[password] {
		suggestions = append(suggestions, 
			"Choose a more unique password - this one is too common")
	}

	return suggestions
}