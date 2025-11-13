package service

import (
	"context"
	"time"

	"github.com/denisakp/pulseguard/pkg/jwt"
)

// AuthService handles authentication logic
type AuthService struct {
	email       string
	password    string
	jwtManager  *jwt.Manager
}

// NewAuthService creates a new authentication service with hardcoded credentials
func NewAuthService(email, password, jwtSecret string) *AuthService {
	// JWT token valid for 24 hours
	jwtManager := jwt.NewManager(jwtSecret, "pulseguard", 24*time.Hour)
	
	return &AuthService{
		email:      email,
		password:   password,
		jwtManager: jwtManager,
	}
}

// Login validates credentials and returns a JWT token
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	// Validate credentials
	if email != s.email || password != s.password {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtManager.Generate(email)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken validates a JWT token and returns the email
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	claims, err := s.jwtManager.Validate(tokenString)
	if err != nil {
		return "", ErrInvalidToken
	}

	return claims.Email, nil
}
