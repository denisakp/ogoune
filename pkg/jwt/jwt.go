package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Claims represents the JWT claims structure
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Manager handles JWT token generation and validation
type Manager struct {
	secretKey []byte
	issuer    string
	duration  time.Duration
}

// NewManager creates a new JWT manager with the given secret key
func NewManager(secretKey string, issuer string, duration time.Duration) *Manager {
	return &Manager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		duration:  duration,
	}
}

// Generate creates a new JWT token for the given email
func (m *Manager) Generate(email string) (string, error) {
	now := time.Now()
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Validate validates a JWT token and returns the claims
func (m *Manager) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
