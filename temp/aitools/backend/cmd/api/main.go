package main

import (
	"log"
	"net/http"
	"os"

	"backend/internal/auth"
	"backend/internal/chat"
	"backend/internal/db"
	"backend/internal/rag"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	router := mux.NewRouter()

	// 配置CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"},
		AllowCredentials: false,
		ExposedHeaders:   []string{"Content-Length"},
	})
	router.Use(corsMiddleware.Handler)

	// Auth routes
	router.HandleFunc("/api/register", auth.RegisterHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/login", auth.LoginHandler).Methods("POST", "OPTIONS")

	// Password change route with JWT middleware
	passwordRouter := router.PathPrefix("/api/user").Subrouter()
	passwordRouter.Use(auth.JWTMiddleware)
	passwordRouter.HandleFunc("/change-password", auth.ChangePasswordHandler).Methods("POST", "OPTIONS")

	// Chat routes with JWT middleware
	chatRouter := router.PathPrefix("/api/chat").Subrouter()
	chatRouter.Use(auth.JWTMiddleware)

	chatRouter.HandleFunc("/history", chat.GetChatHistoryHandler).Methods("GET", "OPTIONS")
	chatRouter.HandleFunc("/new", chat.CreateChatHandler).Methods("POST", "OPTIONS")
	chatRouter.HandleFunc("/{id}/messages", chat.GetChatMessagesHandler).Methods("GET", "OPTIONS")
	chatRouter.HandleFunc("/{id}/messages", chat.SendMessageHandler).Methods("POST", "OPTIONS")
	chatRouter.HandleFunc("/{id}/messages/stream", chat.SendMessageStreamHandler).Methods("GET", "POST", "OPTIONS")
	chatRouter.HandleFunc("/{id}/messages/ai", chat.SaveAIMessageHandler).Methods("POST", "OPTIONS")
	chatRouter.HandleFunc("/{id}/title", chat.UpdateChatTitleHandler).Methods("PUT", "OPTIONS")
	chatRouter.HandleFunc("/{id}", chat.DeleteChatHandler).Methods("DELETE", "OPTIONS")
	chatRouter.HandleFunc("/{id}/info", chat.GetChatInfoHandler).Methods("GET", "OPTIONS")
	chatRouter.HandleFunc("/{id}/model", chat.UpdateChatModelHandler).Methods("PUT", "OPTIONS")
	chatRouter.HandleFunc("/models", chat.GetAvailableModelsHandler).Methods("GET", "OPTIONS")

	// RAG routes
	ragRouter := router.PathPrefix("/api/rag").Subrouter()
	ragRouter.Use(auth.JWTMiddleware)

	ragRouter.HandleFunc("/upload", rag.UploadHandler).Methods("POST", "OPTIONS")
	ragRouter.HandleFunc("/documents", rag.ListDocumentsHandler).Methods("GET", "OPTIONS")
	ragRouter.HandleFunc("/document/{doc_id}", rag.DeleteDocumentHandler).Methods("DELETE", "OPTIONS")
	ragRouter.HandleFunc("/document/{doc_id}/reprocess", rag.ReprocessDocumentHandler).Methods("POST", "OPTIONS")
	ragRouter.HandleFunc("/query", rag.QueryHandler).Methods("POST", "OPTIONS")
	ragRouter.HandleFunc("/status/{task_id}", rag.GetStatusHandler).Methods("GET", "OPTIONS")
	ragRouter.HandleFunc("/clear-vectors", rag.ClearVectorDBHandler).Methods("POST", "OPTIONS")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is running on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
