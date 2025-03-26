package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"rag-backend/internal/app"
	"rag-backend/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Handle clearing vector database request
func HandleClearVectors(c *gin.Context) {
	err := database.ClearQdrantCollection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to clear vector database: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vector database has been cleared"})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or failed to load, will use default values")
	}

	// Initialize MongoDB and Qdrant connections
	database.ConnectMongoDB()
	database.ConnectQdrant()

	// Create Gin router
	router := gin.Default()

	// Set file upload size limit
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Register routes
	// Document upload
	router.POST("/upload", app.HandleUpload)

	// Get document list
	router.GET("/documents", app.HandleListDocuments)

	// Delete document
	router.DELETE("/document/:doc_id", app.HandleDeleteDocument)

	// Clear all documents
	router.POST("/clear-all-documents", app.HandleClearAllDocuments)

	// Clean up invalid document records
	router.POST("/cleanup-invalid-documents", app.HandleCleanupInvalidDocuments)

	// Reprocess document
	router.POST("/document/:doc_id/reprocess", app.HandleReprocessDocument)

	// 使用多Agent系统处理文档
	router.POST("/document/:doc_id/multi-agent", app.HandleMultiAgentProcess)

	// Query
	router.POST("/query", app.HandleQuery)

	// Get document status
	router.GET("/status/:task_id", app.HandleStatus)

	// Clear vector database
	router.POST("/clear-vectors", HandleClearVectors)

	// Get port number
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081" // Default port
	}

	fmt.Printf("✅ Server started successfully, listening on port %s\n", port)
	log.Fatal(router.Run(":" + port))
}
