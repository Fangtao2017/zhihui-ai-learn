package app

import (
	"context"
	"fmt"
	"net/http"

	"rag-backend/internal/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// Define source document structure
type SourceDocument struct {
	DocumentID   string `json:"document_id"`
	DocumentName string `json:"document_name"`
	Content      string `json:"content"`
}

// HandleQuery processes RAG queries
func HandleQuery(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if query is empty
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
		return
	}

	// Print debug information
	fmt.Printf("Received query request: %s\n", req.Query)

	// Generate query vector
	vector, err := database.GetEmbedding(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vector: " + err.Error()})
		return
	}

	// Query Qdrant
	searchResults, err := database.SearchQdrantWithMetadata(vector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed: " + err.Error()})
		return
	}

	// Extract text content for generating answers
	var textResults []string
	var sources []SourceDocument

	// Print debug information
	fmt.Printf("Query: %s\n", req.Query)
	fmt.Printf("Number of search results: %d\n", len(searchResults))

	// If no related content was found
	if len(searchResults) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"answer":  "Sorry, I couldn't find any information related to your question.",
			"sources": []SourceDocument{},
		})
		return
	}

	// Check if the search results have high enough similarity
	// Set a high similarity threshold to determine if results are truly relevant
	const relevanceThreshold = 0.65
	var hasRelevantResults = false

	// Check if any result has similarity above the threshold
	for _, result := range searchResults {
		if result.Score > relevanceThreshold {
			hasRelevantResults = true
			break
		}
	}

	// If all results have similarity below the threshold, consider no relevant content found
	if !hasRelevantResults {
		fmt.Printf("All search results have similarity below threshold %.2f, no relevant content found\n", relevanceThreshold)
		c.JSON(http.StatusOK, gin.H{
			"answer":  "Sorry, I don't have information related to your question in my knowledge base. This question might be outside my expertise.",
			"sources": []SourceDocument{},
		})
		return
	}

	// Use all search results, no filtering
	for _, result := range searchResults {
		textResults = append(textResults, result.Text)

		// Get document name - improved version
		// First set a default name using the last few characters of the document ID
		docName := "Document-" + result.DocID[len(result.DocID)-8:]

		if result.DocID != "" {
			// Get document information from MongoDB
			var doc bson.M
			err := database.MongoCollection.FindOne(
				context.Background(),
				bson.M{"_id": result.DocID},
			).Decode(&doc)

			if err == nil && doc != nil && doc["name"] != nil {
				// If document is found and has a name field, use it
				docName = doc["name"].(string)
			} else {
				fmt.Printf("Using default document name, Document ID: %s\n", result.DocID)
			}
		}

		// Create source document information
		source := SourceDocument{
			DocumentID:   result.DocID,
			DocumentName: docName,
			Content:      result.Text,
		}

		sources = append(sources, source)
	}

	// Call OpenAI to generate answer
	fmt.Printf("Preparing to call GenerateResponse\n")
	answer, err := database.GenerateResponse(textResults, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate answer: " + err.Error()})
		return
	}

	fmt.Printf("Generated answer length: %d characters\n", len(answer))
	fmt.Printf("Answer preview: %s...\n", truncateString(answer, 100))

	// Return answer and sources
	c.JSON(http.StatusOK, gin.H{
		"answer":  answer,
		"sources": sources,
	})
}

// Helper function: truncate string
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
