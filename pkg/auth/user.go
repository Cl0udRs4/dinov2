package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// UserAuth represents user authentication configuration
type UserAuth struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash,omitempty"`
	Password     string `json:"password,omitempty"` // Only used for configuration, not stored in memory
	Role         string `json:"role"`
}

// Global user state
var userState *UserAuth

// SetUserAuth sets the user authentication configuration
func SetUserAuth(auth *UserAuth) {
	fmt.Printf("Setting user auth: %+v\n", auth)
	userState = auth
}

// GetUserAuth returns the user authentication configuration
func GetUserAuth() *UserAuth {
	return userState
}

// ValidateUserCredentials validates user credentials
func ValidateUserCredentials(username, password string) (string, bool) {
	fmt.Printf("Validating credentials: username=%s, password=%s\n", username, password)
	
	if userState == nil {
		fmt.Println("User state is nil")
		return "", false
	}
	
	fmt.Printf("Stored username: %s, passwordHash: %s\n", userState.Username, userState.PasswordHash)
	
	// Check if username matches
	if username != userState.Username {
		fmt.Println("Username does not match")
		return "", false
	}
	
	// Check if password matches
	if err := bcrypt.CompareHashAndPassword([]byte(userState.PasswordHash), []byte(password)); err != nil {
		fmt.Printf("Password does not match: %v\n", err)
		return "", false
	}
	
	fmt.Println("Credentials valid, returning role:", userState.Role)
	// Return the user's role
	return userState.Role, true
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
