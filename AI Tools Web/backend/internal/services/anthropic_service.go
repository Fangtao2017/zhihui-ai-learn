package services

import (
	"backend/internal/models"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// AnthropicService implements the LLMService interface for Anthropic models
type AnthropicService struct {
	CurrentModel string
}

// GetModelName returns the name of the current model
func (s *AnthropicService) GetModelName() string {
	return s.CurrentModel
}

// GetModelProvider returns "anthropic" as the provider
func (s *AnthropicService) GetModelProvider() string {
	return "anthropic"
}

// CallModel calls the Anthropic model with a single message
func (s *AnthropicService) CallModel(message string, model string) (string, error) {
	s.CurrentModel = model
	return CallAnthropic(message, model)
}

// CallModelStreamWithHistory calls the Anthropic model with streaming and message history
func (s *AnthropicService) CallModelStreamWithHistory(w http.ResponseWriter, message string, model string, messages []models.Message) error {
	s.CurrentModel = model
	return CallAnthropicStreamWithHistory(w, message, model, messages)
}

// CallAnthropic calls the Anthropic API to get a response
func CallAnthropic(message string, model string) (string, error) {
	// 检查模型别名，如果存在映射关系则使用映射后的正式模型名称
	if mappedModel, ok := models.ModelAliases[model]; ok {
		log.Printf("Mapping model from %s to %s", model, mappedModel)
		model = mappedModel
	} else if strings.Contains(model, "claude") && !strings.Contains(model, "-20") {
		// 如果是Claude模型但不包含日期后缀，检查是否需要添加日期后缀
		if strings.Contains(model, "claude-3-5-sonnet") || strings.Contains(model, "Claude 3.5 Sonnet") {
			actualModel := "claude-3-5-sonnet-20241022" // 使用最新版本
			log.Printf("自动使用最新版本Claude Sonnet模型: %s -> %s", model, actualModel)
			model = actualModel
		} else if strings.Contains(model, "claude-3-opus") || strings.Contains(model, "Claude 3 Opus") {
			actualModel := "claude-3-opus-20240229"
			log.Printf("自动添加日期后缀到Claude Opus模型: %s -> %s", model, actualModel)
			model = actualModel
		}
	}

	// 尝试从环境变量获取API密钥
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")

	// 如果环境变量中没有API密钥，尝试从配置文件读取
	if apiKey == "" {
		// 尝试从当前目录读取 .env 文件
		if envContent, err := os.ReadFile(".env"); err == nil {
			lines := strings.Split(string(envContent), "\n")
			for _, line := range lines {
				// 跳过注释
				if strings.HasPrefix(strings.TrimSpace(line), "#") {
					continue
				}
				// 解析ANTHROPIC_API_KEY
				if strings.Contains(line, "ANTHROPIC_API_KEY=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						apiKey = strings.TrimSpace(parts[1])
						log.Printf("Found API key in .env file")
					}
				}
				// 解析ANTHROPIC_BASE_URL
				if strings.Contains(line, "ANTHROPIC_BASE_URL=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						baseURL = strings.TrimSpace(parts[1])
					}
				}
			}
		}

		// 尝试从常见的配置文件位置读取
		configLocations := []string{
			".env",
			"../../.env",
			"../../../.env",
			"config.env",
			"../../config.env",
			"../../../config.env",
			"config/config.env",
			"../config/config.env",
			"../../config/config.env",
		}

		for _, location := range configLocations {
			if apiKey != "" {
				break // 如果已经找到API密钥，就跳出循环
			}
			if envContent, err := os.ReadFile(location); err == nil {
				lines := strings.Split(string(envContent), "\n")
				for _, line := range lines {
					// 跳过注释和空行
					if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
						continue
					}
					// 解析ANTHROPIC_API_KEY
					if strings.Contains(line, "ANTHROPIC_API_KEY=") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							apiKey = strings.TrimSpace(parts[1])
							log.Printf("Found API key in %s", location)
						}
					}
					// 解析ANTHROPIC_BASE_URL
					if strings.Contains(line, "ANTHROPIC_BASE_URL=") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							baseURL = strings.TrimSpace(parts[1])
						}
					}
				}
			}
		}
	}

	// 如果仍然无法找到API密钥，使用硬编码的密钥
	if apiKey == "" {
		log.Println("Using hardcoded API key as last resort")
		apiKey = "sk-ant-api03-EMmzA6LRB49QxabXcOR_PU1WWiyWaVTE3i-Tp1t21V0iHY9fHUUteP5M4jkkK"
	}

	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	// Build request with model identity information added to the prompt
	enhancedMessage := fmt.Sprintf("You are %s, an AI assistant. When asked about your identity or model name, explicitly identify yourself as %s.\n\nUser question: %s", model, model, message)

	anthropicMessages := []map[string]string{
		{
			"role":    "user",
			"content": enhancedMessage,
		},
	}

	requestBody := map[string]interface{}{
		"model":      model,
		"messages":   anthropicMessages,
		"max_tokens": 2000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	url := baseURL + "/v1/messages"
	log.Printf("Making request to Anthropic API: %s", url)

	// 记录完整请求体用于调试
	log.Printf("Complete Anthropic API request body: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("创建请求错误: %v", err)
		return "", fmt.Errorf("创建请求错误: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("API错误 (状态码 %d): %s", resp.StatusCode, string(body))
		log.Printf("完整的响应头: %v", resp.Header)
		return "", fmt.Errorf("API错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	var response models.AnthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v", err)
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	// Extract the text content from response
	var fullContent string
	for _, content := range response.Content {
		if content.Type == "text" {
			fullContent += content.Text
		}
	}

	if fullContent == "" {
		log.Println("No response from API")
		return "", errors.New("no response from API")
	}

	log.Println("Successfully received response from Anthropic API")
	return fullContent, nil
}

// CallAnthropicStreamWithHistory uses streaming response to call Anthropic API with message history
func CallAnthropicStreamWithHistory(w http.ResponseWriter, message string, model string, messages []models.Message) error {
	// 检查模型别名，如果存在映射关系则使用映射后的正式模型名称
	if mappedModel, ok := models.ModelAliases[model]; ok {
		log.Printf("Mapping model from %s to %s", model, mappedModel)
		model = mappedModel
	} else if strings.Contains(model, "claude") && !strings.Contains(model, "-20") {
		// 如果是Claude模型但不包含日期后缀，检查是否需要添加日期后缀
		if strings.Contains(model, "claude-3-5-sonnet") || strings.Contains(model, "Claude 3.5 Sonnet") {
			actualModel := "claude-3-5-sonnet-20241022" // 使用最新版本
			log.Printf("自动使用最新版本Claude Sonnet模型: %s -> %s", model, actualModel)
			model = actualModel
		} else if strings.Contains(model, "claude-3-opus") || strings.Contains(model, "Claude 3 Opus") {
			actualModel := "claude-3-opus-20240229"
			log.Printf("自动添加日期后缀到Claude Opus模型: %s -> %s", model, actualModel)
			model = actualModel
		}
	}

	// 尝试从环境变量获取API密钥
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")

	// 如果环境变量中没有API密钥，尝试从配置文件读取
	if apiKey == "" {
		// 尝试从当前目录读取 .env 文件
		if envContent, err := os.ReadFile(".env"); err == nil {
			lines := strings.Split(string(envContent), "\n")
			for _, line := range lines {
				// 跳过注释
				if strings.HasPrefix(strings.TrimSpace(line), "#") {
					continue
				}
				// 解析ANTHROPIC_API_KEY
				if strings.Contains(line, "ANTHROPIC_API_KEY=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						apiKey = strings.TrimSpace(parts[1])
						log.Printf("Found API key in .env file for streaming")
					}
				}
				// 解析ANTHROPIC_BASE_URL
				if strings.Contains(line, "ANTHROPIC_BASE_URL=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						baseURL = strings.TrimSpace(parts[1])
					}
				}
			}
		}

		// 尝试从常见的配置文件位置读取
		configLocations := []string{
			".env",
			"../../.env",
			"../../../.env",
			"config.env",
			"../../config.env",
			"../../../config.env",
			"config/config.env",
			"../config/config.env",
			"../../config/config.env",
		}

		for _, location := range configLocations {
			if apiKey != "" {
				break // 如果已经找到API密钥，就跳出循环
			}
			if envContent, err := os.ReadFile(location); err == nil {
				lines := strings.Split(string(envContent), "\n")
				for _, line := range lines {
					// 跳过注释和空行
					if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
						continue
					}
					// 解析ANTHROPIC_API_KEY
					if strings.Contains(line, "ANTHROPIC_API_KEY=") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							apiKey = strings.TrimSpace(parts[1])
							log.Printf("Found API key in %s for streaming", location)
						}
					}
					// 解析ANTHROPIC_BASE_URL
					if strings.Contains(line, "ANTHROPIC_BASE_URL=") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							baseURL = strings.TrimSpace(parts[1])
						}
					}
				}
			}
		}
	}

	// 如果仍然无法找到API密钥，使用硬编码的密钥
	if apiKey == "" {
		log.Println("Using hardcoded API key as last resort for streaming")
		apiKey = "sk-ant-api03-EMmzA6LRB49QxabXcOR_PU1WWiyWaVTE3i-Tp1t21V0iHY9fHUUteP5M4jkkK"
	}

	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	log.Printf("Starting stream request to Anthropic with model: %s and %d messages", model, len(messages))

	// 1. Process system prompt
	var systemPrompt string
	var anthropicMessages []map[string]string

	// Log all messages for debugging
	for i, msg := range messages {
		contentPreview := msg.Content
		if len(contentPreview) > 50 {
			contentPreview = contentPreview[:50] + "..."
		}
		log.Printf("Message %d: Role=%s, Content=%s", i, msg.Role, contentPreview)
	}

	// Extract system prompt
	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			break
		}
	}

	// 2. Build message array
	for _, msg := range messages {
		// Skip system messages, handled separately
		if msg.Role == "system" {
			continue
		}

		// Claude API supports user and assistant roles
		if msg.Role == "user" || msg.Role == "assistant" {
			anthropicMessages = append(anthropicMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	// 3. Build API request
	requestData := map[string]interface{}{
		"model":      model,
		"stream":     true,
		"max_tokens": 4000,
	}

	// Add messages
	if len(anthropicMessages) > 0 {
		requestData["messages"] = anthropicMessages
	}

	// Add system prompt
	if systemPrompt != "" {
		// Enhance system prompt with model identity
		enhancedSystemPrompt := fmt.Sprintf("%s\n\nYou are %s. When asked about your identity or abilities, accurately identify yourself as %s.",
			systemPrompt, model, model)
		requestData["system"] = enhancedSystemPrompt
	}

	// 4. Serialize request data
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return err
	}

	// Log request for debugging
	log.Printf("Complete Anthropic API request body: %s", string(jsonData))

	// 5. Create HTTP request
	req, err := http.NewRequest("POST", baseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	// 6. Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Accept", "text/event-stream")

	// 7. Create HTTP client and send request
	client := &http.Client{
		Timeout: 300 * time.Second, // Increased timeout
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		fmt.Fprintf(w, "data: ERROR: Request failed: %v\n\n", err)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return err
	}
	defer resp.Body.Close()

	// 8. Check response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
		log.Printf(errorMsg)
		fmt.Fprintf(w, "data: ERROR: %s\n\n", errorMsg)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return errors.New(errorMsg)
	}

	// Log response headers for debugging
	log.Printf("Anthropic API response status: %d", resp.StatusCode)
	log.Printf("Anthropic API response headers: %v", resp.Header)

	// 9. Process streaming response
	reader := bufio.NewReader(resp.Body)
	contentBuffer := ""
	eventCount := 0
	lastEventTime := time.Now()

	// 不发送原始的stream_start事件，这会导致前端显示JSON
	// 改为只发送一个空的数据包，确认连接已建立
	fmt.Fprintf(w, "data: \n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Stream ended (EOF)")
				break
			}
			log.Printf("Error reading stream: %v", err)
			fmt.Fprintf(w, "data: ERROR: Failed to read stream: %v\n\n", err)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return err
		}

		now := time.Now()
		sinceLastEvent := now.Sub(lastEventTime).Seconds()

		// Log if no events for a while
		if sinceLastEvent > 5 {
			log.Printf("No events for %.1f seconds...", sinceLastEvent)
		}

		lastEventTime = now

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Log raw response line for debug
		log.Printf("Raw SSE line: %s", line)

		// Parse event type and data
		var data string

		if strings.HasPrefix(line, "event:") {
			eventType := strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			log.Printf("Event type: %s", eventType)
			continue // Read next line for data
		}

		if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			// Special end marker
			if data == "[DONE]" {
				log.Printf("Received [DONE] marker")
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				break
			}

			// Process JSON data
			var responseChunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &responseChunk); err != nil {
				log.Printf("Error parsing chunk: %v, data: %s", err, data)
				continue
			}

			// Process data by event type
			eventCount++
			eventType, _ := responseChunk["type"].(string)
			log.Printf("Processing event type: %s", eventType)

			// Extract content based on event type
			var contentText string
			var contentAdded bool

			// Check for content_block_delta events (newer API response format)
			if eventType == "content_block_delta" {
				if delta, ok := responseChunk["delta"].(map[string]interface{}); ok {
					if text, ok := delta["text"].(string); ok && text != "" {
						contentText = text
						contentBuffer += text
						contentAdded = true
						log.Printf("Content text from content_block_delta: %s", text)
					}
				}
			}

			// Method 1: Check delta.text (standard streaming format)
			if !contentAdded && responseChunk["delta"] != nil {
				if delta, ok := responseChunk["delta"].(map[string]interface{}); ok {
					if text, ok := delta["text"].(string); ok && text != "" {
						contentText = text
						contentBuffer += text
						contentAdded = true
						log.Printf("Delta text: %s", text)
					}
				}
			}

			// Method 2: Check content[0].text (non-streaming format)
			if !contentAdded && responseChunk["content"] != nil {
				if content, ok := responseChunk["content"].([]interface{}); ok && len(content) > 0 {
					if contentItem, ok := content[0].(map[string]interface{}); ok {
						if text, ok := contentItem["text"].(string); ok && text != "" {
							contentText = text
							contentBuffer += text
							contentAdded = true
							log.Printf("Content text: %s", text)
						}
					}
				}
			}

			// 不再直接传递原始JSON响应
			if !contentAdded {
				log.Printf("未能从数据中提取文本，尝试从其他字段提取: %v", responseChunk)
				// 不向客户端发送任何内容，只记录日志
				continue
			}

			// 直接发送提取的文本内容，不再包装成JSON
			if contentText != "" {
				fmt.Fprintf(w, "data: %s\n\n", contentText)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	}

	log.Printf("Stream completed: processed %d events, total response length: %d characters",
		eventCount, len(contentBuffer))

	// If no content received, send error message
	if contentBuffer == "" {
		log.Printf("WARNING: No content received from Anthropic API after %d events", eventCount)

		// Try to re-encode and send request body for debugging
		debugRequestBody, _ := json.MarshalIndent(requestData, "", "  ")
		log.Printf("Debug request body: %s", string(debugRequestBody))

		// Send error to client
		errorMsg := "No content received from Anthropic API"
		fmt.Fprintf(w, "data: ERROR: %s\n\n", errorMsg)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Ensure [DONE] marker is sent
		fmt.Fprintf(w, "data: [DONE]\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		return errors.New(errorMsg)
	}

	// 不再发送完整的最终响应，避免内容重复
	log.Printf("Stream complete, sending DONE marker")

	// 只发送流结束标记
	fmt.Fprintf(w, "data: [DONE]\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
