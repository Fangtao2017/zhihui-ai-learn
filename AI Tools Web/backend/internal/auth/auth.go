package auth

import (
	"backend/internal/db"
	"backend/internal/models"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	// "github.com/gorilla/mux"
)

// logs registration request
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received registration request")

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to register user: %+v", user)

	collection := db.GetCollection(db.UserCollection)

	// logs check username and email alr exist
	var existingUser models.User
	err := collection.FindOne(context.TODO(), bson.M{"$or": []interface{}{
		bson.M{"username": user.Username},
		bson.M{"email": user.Email},
	}}).Decode(&existingUser)

	if err != mongo.ErrNoDocuments {
		log.Printf("User already exists - username: %s, email: %s", user.Username, user.Email) // 新增
		http.Error(w, "Username or Email already exists", http.StatusConflict)
		return
	}

	// insert new user
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully registered user with ID: %v", result.InsertedID)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

// logs Login request
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received login request")

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := db.GetCollection(db.UserCollection)

	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{
		"email":    credentials.Email,
		"password": credentials.Password,
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Login failed: invalid credentials for email: %s", credentials.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		log.Printf("Database error during login: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	token, err := GenerateToken(user.Username, user.Email)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// logs Login successful(terminal)
	log.Printf("User logged in successfully: %s", user.Email)
	log.Printf("User %s logged in with JWT token: %s", user.Email, token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Login successful",
		"username": user.Username,
		"token":    token,
	})
}

// ChangePasswordHandler handles password change requests
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received password change request")

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get user information from token
	user, err := GetUserFromRequest(r)
	if err != nil {
		log.Printf("Failed to get user from token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var passwordChange struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&passwordChange); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from database to verify old password
	collection := db.GetCollection(db.UserCollection)

	var dbUser models.User
	err = collection.FindOne(context.TODO(), bson.M{
		"email": user.Email,
	}).Decode(&dbUser)

	if err != nil {
		log.Printf("Failed to find user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify old password
	if dbUser.Password != passwordChange.OldPassword {
		log.Printf("Invalid old password for user: %s", user.Email)
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	// Update password
	update := bson.M{
		"$set": bson.M{
			"password": passwordChange.NewPassword,
		},
	}

	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"email": user.Email},
		update,
	)

	if err != nil {
		log.Printf("Failed to update password: %v", err)
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	log.Printf("Password updated successfully for user: %s", user.Email)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully",
	})
}
