package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"backend/internal/models"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request method: %v", r.Header.Get("Authorization"))
		// 处理 OPTIONS 请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		log.Printf("Bearer token: %v", bearerToken[1])
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token := bearerToken[1]
		claims, err := ValidateToken(token)
		log.Printf("Claims: %v", claims)
		if err != nil {
			// 更详细的错误信息
			if err.Error() == "Token is expired" {
				log.Printf("Token has expired")
				http.Error(w, "Token has expired", http.StatusUnauthorized)
			} else {
				log.Printf("Invalid token")

				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
			log.Printf("Error: %v", err)
			return
		}

		// 将用户信息存入上下文
		ctx := context.WithValue(r.Context(), "user", claims)
		log.Printf("Context: %v", ctx.Value("user"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromRequest extracts user information from the JWT token in the request
func GetUserFromRequest(r *http.Request) (*models.User, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return nil, fmt.Errorf("no authorization token provided")
	}

	// Remove 'Bearer ' prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Parse and validate the token
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Create user from claims
	user := &models.User{
		Username: claims.Username,
		Email:    claims.Email,
	}

	return user, nil
}
