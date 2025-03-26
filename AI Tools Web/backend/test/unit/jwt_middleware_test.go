package auth_test

import (
	"backend/internal/auth"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

func init() {
	// 为测试设置环境变量
	os.Setenv("JWT_SECRET_KEY", "test_secret_key")
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "Generate token normally",
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			email:    "test@example.com",
			wantErr:  true,
		},
		{
			name:     "empty email",
			username: "testuser",
			email:    "",
			wantErr:  true,
		},
		{
			name:     "Both empty",
			username: "",
			email:    "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.username, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() The returned token should not be empty")
			}
			if tt.wantErr && err == nil {
				t.Error("GenerateToken() Should return an error, but doesn't")
			}
		})
	}
}

func generateTokenWithExpiration(username, email string, duration time.Duration) (string, error) {
	claims := &auth.Claims{
		Username: username,
		Email:    email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

func TestValidateToken(t *testing.T) {
	username := "testuser"
	email := "test@example.com"

	validToken, _ := auth.GenerateToken(username, email)
	expiredToken, _ := generateTokenWithExpiration(username, email, -time.Hour)

	tests := []struct {
		name         string
		token        string
		wantUsername string
		wantEmail    string
		wantErr      bool
	}{
		{
			name:         "Valid token",
			token:        validToken,
			wantUsername: username,
			wantEmail:    email,
			wantErr:      false,
		},
		{
			name:         "Expired token",
			token:        expiredToken,
			wantUsername: "",
			wantEmail:    "",
			wantErr:      true,
		},
		{
			name:         "Invalid token",
			token:        "invalid.token.here",
			wantUsername: "",
			wantEmail:    "",
			wantErr:      true,
		},
		{
			name:         "empty token",
			token:        "",
			wantUsername: "",
			wantEmail:    "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.Username != tt.wantUsername {
					t.Errorf("ValidateToken() username = %v, want %v", claims.Username, tt.wantUsername)
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("ValidateToken() email = %v, want %v", claims.Email, tt.wantEmail)
				}
			}
		})
	}
}
