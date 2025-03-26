package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"rag-backend/internal/database"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AgentResult represents the processing result of each Agent
type AgentResult struct {
	Content map[string]interface{} `json:"content"`
	Status  string                 `json:"status"`
	Error   string                 `json:"error,omitempty"`
}

// MultiAgentResult represents the overall result of the multi-Agent system
type MultiAgentResult struct {
	DocID           string      `json:"doc_id"`
	FileName        string      `json:"file_name"`
	ProcessedAt     time.Time   `json:"processed_at"`
	ContentAgent    AgentResult `json:"content_agent"`   // Content Analysis Agent
	KnowledgeAgent  AgentResult `json:"knowledge_agent"` // Knowledge Extraction Agent
	SummaryAgent    AgentResult `json:"summary_agent"`   // Summary Agent
	FormatAgent     AgentResult `json:"format_agent"`    // Formatting Agent
	Status          string      `json:"status"`
	Error           string      `json:"error,omitempty"`
	MarkdownContent string      `json:"markdown_content"` // Markdown formatted content
	AnkiCards       interface{} `json:"anki_cards"`       // Anki cards data
}

// HandleMultiAgentProcess processes multi-Agent system API requests
func HandleMultiAgentProcess(c *gin.Context) {
	docID := c.Param("doc_id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID cannot be empty"})
		return
	}

	// Query document information
	objID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID format"})
		return
	}

	// Query document
	var doc bson.M
	err = database.MongoCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Print document information for debugging
	fmt.Printf("Document information: %+v\n", doc)

	// Check each field
	fmt.Printf("ID: %v\n", doc["_id"])
	fmt.Printf("name: %v\n", doc["name"])
	fmt.Printf("status: %v\n", doc["status"])
	fmt.Printf("chunks: %v\n", doc["chunks"])

	// Get filename
	fileName, ok := doc["name"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File name format error"})
		return
	}

	// Get document content from chunks for analysis
	var docContent string
	var chunkTexts []string
	if chunks, ok := doc["chunks"].(primitive.A); ok {
		for _, chunk := range chunks {
			if chunkDoc, ok := chunk.(bson.M); ok {
				if content, ok := chunkDoc["content"].(string); ok {
					docContent += content + "\n\n"
					chunkTexts = append(chunkTexts, content)
				}
			}
		}
	}

	if docContent == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get document content"})
		return
	}

	fmt.Printf("Document content retrieved, length: %d characters, total %d chunks\n", len(docContent), len(chunkTexts))

	// Start processing time
	startTime := time.Now()

	// Process using LangChain multi-Agent system
	fmt.Println("[Multi-Agent System] Starting to process document...")

	// 1. Content Analysis Agent - Extract document structure
	fmt.Println("[Multi-Agent System] Executing Content Analysis Agent...")
	contentAgentResult, err := runLangChainContentAnalysisAgent(docContent)
	if err != nil {
		fmt.Printf("[Multi-Agent System] Content analysis failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Content analysis failed: %v", err)})
		return
	}

	// Get chapters from results
	chaptersInterface, ok := contentAgentResult["chapters"]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Content analysis did not return chapter structure"})
		return
	}

	chaptersResult, ok := chaptersInterface.([]map[string]interface{})
	if !ok {
		// Try conversion
		chaptersArray, ok := chaptersInterface.([]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Chapter structure format error"})
			return
		}

		// Convert []interface{} to []map[string]interface{}
		chaptersResult = make([]map[string]interface{}, len(chaptersArray))
		for i, v := range chaptersArray {
			if m, ok := v.(map[string]interface{}); ok {
				chaptersResult[i] = m
			} else {
				// If unable to convert, create a default chapter
				chaptersResult[i] = map[string]interface{}{
					"title":   fmt.Sprintf("Chapter %d", i+1),
					"level":   1,
					"content": "Chapter content",
				}
			}
		}
	}

	// 2. Knowledge Extraction Agent - Extract key concepts
	fmt.Println("[Multi-Agent System] Executing Knowledge Extraction Agent...")
	knowledgeAgentResult, err := runLangChainKnowledgeExtractionAgent(chaptersResult)
	if err != nil {
		fmt.Printf("[Multi-Agent System] Knowledge extraction failed: %v\n", err)
		// Continue execution, don't interrupt processing
	}

	// Get key concepts
	var conceptsResult []map[string]interface{}
	if err == nil && knowledgeAgentResult != nil {
		if concepts, ok := knowledgeAgentResult["key_concepts"]; ok {
			if conceptArray, ok := concepts.([]interface{}); ok {
				// Convert
				conceptsResult = make([]map[string]interface{}, len(conceptArray))
				for i, v := range conceptArray {
					if m, ok := v.(map[string]interface{}); ok {
						conceptsResult[i] = m
					}
				}
			} else if conceptMaps, ok := concepts.([]map[string]interface{}); ok {
				conceptsResult = conceptMaps
			}
		}
	}

	// 3. Summary Generation Agent - Generate summaries
	fmt.Println("[Multi-Agent System] Executing Summary Generation Agent...")
	summaryAgentResult, err := runLangChainSummaryAgent(chaptersResult)
	if err != nil {
		fmt.Printf("[Multi-Agent System] Summary generation failed: %v\n", err)
		// Continue execution, don't interrupt processing
	}

	// Get summaries
	var summariesResult []map[string]interface{}
	if err == nil && summaryAgentResult != nil {
		if summaries, ok := summaryAgentResult["summaries"]; ok {
			if summaryArray, ok := summaries.([]interface{}); ok {
				// Convert
				summariesResult = make([]map[string]interface{}, len(summaryArray))
				for i, v := range summaryArray {
					if m, ok := v.(map[string]interface{}); ok {
						summariesResult[i] = m
					}
				}
			} else if summaryMaps, ok := summaries.([]map[string]interface{}); ok {
				summariesResult = summaryMaps
			}
		}
	}

	// 4. Formatting Agent - Generate final output
	fmt.Println("[Multi-Agent System] Executing Formatting Agent...")
	formatResult, err := runLangChainFormatAgent(chaptersResult, conceptsResult, summariesResult)
	if err != nil {
		fmt.Printf("[Multi-Agent System] Formatting failed: %v\n", err)
		// Use default format
		formatResult = map[string]interface{}{
			"markdown_notes": "# " + fileName + "\n\nContent generation failed",
			"anki_cards":     []map[string]string{},
			"format_version": "1.0",
		}
	}

	// Get Markdown content and Anki cards
	markdownContent, _ := formatResult["markdown_notes"].(string)
	ankiCards, _ := formatResult["anki_cards"].([]map[string]string)

	// Create Agent result
	contentAgent := AgentResult{
		Status: "success",
		Content: map[string]interface{}{
			"chapters": chaptersResult,
		},
	}

	knowledgeAgent := AgentResult{
		Status: "success",
		Content: map[string]interface{}{
			"key_concepts": conceptsResult,
		},
	}

	summaryAgent := AgentResult{
		Status: "success",
		Content: map[string]interface{}{
			"summaries": summariesResult,
		},
	}

	formatAgent := AgentResult{
		Status: "success",
		Content: map[string]interface{}{
			"format_version": "1.0",
		},
	}

	// Calculate processing time
	processingTime := time.Since(startTime).Seconds()

	// Build result
	result := map[string]interface{}{
		"doc_id":           docID,
		"file_name":        fileName,
		"processed_at":     time.Now(),
		"status":           "success",
		"processing_time":  processingTime,
		"content_agent":    contentAgent,
		"knowledge_agent":  knowledgeAgent,
		"summary_agent":    summaryAgent,
		"format_agent":     formatAgent,
		"chapters":         chaptersResult,
		"key_concepts":     conceptsResult,
		"summaries":        summariesResult,
		"markdown_content": markdownContent,
		"anki_cards":       ankiCards,
	}

	fmt.Printf("[Multi-Agent System] Processing completed, time: %.2f seconds\n", processingTime)

	// Return result
	c.JSON(http.StatusOK, result)
}

// Content Analysis Agent processing function
func processWithContentAnalysisAgent(content string) ([]map[string]interface{}, error) {
	// If content is too long, only process the first 16000 characters
	if len(content) > 16000 {
		content = content[:16000]
	}

	// Call OpenAI API to analyze document structure
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Return simulated data
		return []map[string]interface{}{
			{"title": "Introduction", "level": 1, "content": "This is the introduction section of the document..."},
			{"title": "Main Content", "level": 1, "content": "This is the main content of the document..."},
			{"title": "Conclusion", "level": 1, "content": "This is the conclusion section of the document..."},
		}, nil
	}

	// Build API request
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": `You are a document structure analysis expert. You need to identify the chapter structure of the document, including titles, subchapters, and paragraphs.
				Analyze the provided text content and extract its chapter structure, each chapter needs to include a title and content.
				Return JSON format, structure as follows:
				{
					"chapters": [
						{
							"title": "Chapter Title",
							"level": 1,
							"content": "Chapter Content"
						},
						...
					]
				}`,
			},
			{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0.2,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build API request: %v", err)
	}

	// Send API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create API request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()

	// Read API response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %v", err)
	}

	// Check API response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status code: %d, response: %s", resp.StatusCode, string(body))
	}

	// Parse API response
	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %v", err)
	}

	// Extract JSON content
	choices, ok := apiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("API response format error: No choices field")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("API response format error: Unable to get message")
	}

	contentStr, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("API response format error: Unable to get content")
	}

	// Extract JSON part
	jsonStart := strings.Index(contentStr, "{")
	jsonEnd := strings.LastIndex(contentStr, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := contentStr[jsonStart : jsonEnd+1]
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
			return nil, fmt.Errorf("Failed to parse API returned JSON: %v", err)
		}

		// Return chapter data
		if chapters, ok := result["chapters"].([]interface{}); ok {
			chaptersResult := make([]map[string]interface{}, 0, len(chapters))
			for _, ch := range chapters {
				if chMap, ok := ch.(map[string]interface{}); ok {
					chaptersResult = append(chaptersResult, chMap)
				}
			}
			return chaptersResult, nil
		}
		return nil, fmt.Errorf("API response does not contain chapters field")
	}

	return nil, fmt.Errorf("Unable to extract valid JSON structure from API response")
}

// Knowledge Extraction Agent processing function
func processWithKnowledgeExtractionAgent(chapters []map[string]interface{}, fullContent string) ([]map[string]interface{}, error) {
	// If there is no API key, return simulated data
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return []map[string]interface{}{
			{"title": "DSRC", "description": "Dedicated Short Range Communications (Dedicated Short Range Communications) is a technology used for vehicle communication", "chapter": "Introduction"},
			{"title": "IEEE 1609", "description": "IEEE 1609 is the network and security standard for DSRC", "chapter": "Main Content"},
			{"title": "SAE J2735", "description": "SAE J2735 defines the message set and data frame for DSRC", "chapter": "Main Content"},
		}, nil
	}

	// Prepare input content
	var inputContent string
	for _, chapter := range chapters {
		title, _ := chapter["title"].(string)
		content, _ := chapter["content"].(string)

		if len(content) > 2000 {
			content = content[:2000] + "..."
		}

		inputContent += fmt.Sprintf("# %s\n%s\n\n", title, content)
	}

	// Limit overall content length
	if len(inputContent) > 16000 {
		inputContent = inputContent[:16000]
	}

	// Build API request
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": `You are a knowledge extraction expert. Identify and extract key concepts, definitions, formulas, and important knowledge points from the provided document content.
				Return JSON format, structure as follows:
				{
					"key_concepts": [
						{
							"title": "Concept Name",
							"description": "Concept Description",
							"chapter": "Belongs to Chapter"
						},
						...
					]
				}`,
			},
			{
				"role":    "user",
				"content": inputContent,
			},
		},
		"temperature": 0.2,
	}

	// Call API
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build API request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create API request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()

	// Process result
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %v", err)
	}

	// Extract JSON content
	choices, ok := apiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("API response format error: No choices field")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("API response format error: Unable to get message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("API response format error: Unable to get content")
	}

	// Extract JSON part
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := content[jsonStart : jsonEnd+1]
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
			return nil, fmt.Errorf("Failed to parse API returned JSON: %v", err)
		}

		// Return key concepts
		if concepts, ok := result["key_concepts"].([]interface{}); ok {
			conceptsResult := make([]map[string]interface{}, 0, len(concepts))
			for _, c := range concepts {
				if cMap, ok := c.(map[string]interface{}); ok {
					conceptsResult = append(conceptsResult, cMap)
				}
			}
			return conceptsResult, nil
		}
		return nil, fmt.Errorf("API response does not contain key_concepts field")
	}

	return nil, fmt.Errorf("Unable to extract valid JSON structure from API response")
}

// Summary Generation Agent processing function
func processWithSummaryAgent(chapters []map[string]interface{}) ([]map[string]interface{}, error) {
	// If there is no API key, return simulated data
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return []map[string]interface{}{
			{"title": "Introduction", "summary": "This chapter introduces the background and uses of DSRC technology"},
			{"title": "Main Content", "summary": "Detailedly describes the technical implementation and standards of DSRC"},
			{"title": "Conclusion", "summary": "Summarizes the advantages and application prospects of DSRC"},
		}, nil
	}

	// Prepare input content
	var inputContent string
	for _, chapter := range chapters {
		title, _ := chapter["title"].(string)
		content, _ := chapter["content"].(string)

		if len(content) > 2000 {
			content = content[:2000] + "..."
		}

		inputContent += fmt.Sprintf("# %s\n%s\n\n", title, content)
	}

	// Limit overall content length
	if len(inputContent) > 16000 {
		inputContent = inputContent[:16000]
	}

	// Build API request
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": `You are a summary expert. Generate concise summaries for each chapter in the document, highlighting key points.
				Return JSON format, structure as follows:
				{
					"summaries": [
						{
							"title": "Chapter Title",
							"summary": "Chapter Summary"
						},
						...
					]
				}`,
			},
			{
				"role":    "user",
				"content": inputContent,
			},
		},
		"temperature": 0.2,
	}

	// Call API
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build API request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create API request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()

	// Process result
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %v", err)
	}

	// Extract JSON content
	choices, ok := apiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("API response format error: No choices field")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("API response format error: Unable to get message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("API response format error: Unable to get content")
	}

	// Extract JSON part
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := content[jsonStart : jsonEnd+1]
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
			return nil, fmt.Errorf("Failed to parse API returned JSON: %v", err)
		}

		// Return summaries
		if summaries, ok := result["summaries"].([]interface{}); ok {
			summariesResult := make([]map[string]interface{}, 0, len(summaries))
			for _, s := range summaries {
				if sMap, ok := s.(map[string]interface{}); ok {
					summariesResult = append(summariesResult, sMap)
				}
			}
			return summariesResult, nil
		}
		return nil, fmt.Errorf("API response does not contain summaries field")
	}

	return nil, fmt.Errorf("Unable to extract valid JSON structure from API response")
}

// Formatting Agent processing function
func processWithFormatAgent(chapters []map[string]interface{}, concepts []map[string]interface{}, summaries []map[string]interface{}, fileName string) (map[string]interface{}, error) {
	// Generate Markdown format
	var markdownContent strings.Builder

	// Add document title
	markdownContent.WriteString(fmt.Sprintf("# %s Learning Notes\n\n", strings.TrimSuffix(fileName, filepath.Ext(fileName))))
	markdownContent.WriteString("## Table of Contents\n\n")

	// Generate table of contents
	for i, chapter := range chapters {
		title, _ := chapter["title"].(string)
		markdownContent.WriteString(fmt.Sprintf("%d. [%s](#%s)\n", i+1, title, strings.ReplaceAll(strings.ToLower(title), " ", "-")))
	}

	markdownContent.WriteString("\n---\n\n")

	// Generate chapter content
	for i, chapter := range chapters {
		title, _ := chapter["title"].(string)
		content, _ := chapter["content"].(string)

		// Find corresponding summary
		var summary string
		for _, s := range summaries {
			if sTitle, _ := s["title"].(string); sTitle == title {
				summary, _ = s["summary"].(string)
				break
			}
		}

		markdownContent.WriteString(fmt.Sprintf("## %s\n\n", title))

		if summary != "" {
			markdownContent.WriteString("### Summary\n\n")
			markdownContent.WriteString(summary + "\n\n")
		}

		// Find key concepts for this chapter
		conceptCount := 0
		found := false
		for _, concept := range concepts {
			if chap, _ := concept["chapter"].(string); chap == title {
				if !found {
					markdownContent.WriteString("### Key Concepts\n\n")
					found = true
				}

				conceptTitle, _ := concept["title"].(string)
				description, _ := concept["description"].(string)
				markdownContent.WriteString(fmt.Sprintf("- **%s**: %s\n", conceptTitle, description))
				conceptCount++
			}
		}

		if found {
			markdownContent.WriteString("\n")
		}

		// Add selective content display
		if len(content) > 500 {
			// If content is too long, only display a part
			markdownContent.WriteString("### Content Excerpt\n\n")
			markdownContent.WriteString(content[:500] + "...\n\n")
		} else if content != "" {
			markdownContent.WriteString("### Content\n\n")
			markdownContent.WriteString(content + "\n\n")
		}

		// If it's not the last chapter, add separator
		if i < len(chapters)-1 {
			markdownContent.WriteString("---\n\n")
		}
	}

	// Generate Anki flash card content
	var ankiCards []map[string]string

	// Generate Anki cards from key concepts
	for _, concept := range concepts {
		title, _ := concept["title"].(string)
		description, _ := concept["description"].(string)
		chapter, _ := concept["chapter"].(string)

		ankiCards = append(ankiCards, map[string]string{
			"front": fmt.Sprintf("Definition: %s", title),
			"back":  fmt.Sprintf("%s\n\n(From Chapter: %s)", description, chapter),
			"tags":  "Concept, Definition",
		})
	}

	// Generate question and answer cards from summaries
	for _, summary := range summaries {
		title, _ := summary["title"].(string)
		sum, _ := summary["summary"].(string)

		ankiCards = append(ankiCards, map[string]string{
			"front": fmt.Sprintf("Summarize the key points of the \"%s\" chapter", title),
			"back":  sum,
			"tags":  "Summary, Chapter",
		})
	}

	// Build return result
	result := map[string]interface{}{
		"markdown_content": markdownContent.String(),
		"anki_cards":       ankiCards,
		"format_version":   "1.0",
	}

	return result, nil
}
