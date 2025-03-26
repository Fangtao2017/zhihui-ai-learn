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

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// 流式响应的数据结构
type OpenAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// OpenAIService implements the LLMService interface for OpenAI models
type OpenAIService struct {
	CurrentModel string
}

// GetModelName returns the name of the current model
func (s *OpenAIService) GetModelName() string {
	return s.CurrentModel
}

// GetModelProvider returns "openai" as the provider
func (s *OpenAIService) GetModelProvider() string {
	return "openai"
}

// CallModel calls the OpenAI model with a single message
func (s *OpenAIService) CallModel(message string, model string) (string, error) {
	s.CurrentModel = model
	return CallOpenAI(message, model)
}

// CallModelStreamWithHistory calls the OpenAI model with streaming and message history
func (s *OpenAIService) CallModelStreamWithHistory(w http.ResponseWriter, message string, model string, messages []models.Message) error {
	s.CurrentModel = model
	return CallOpenAIStreamWithHistory(w, message, model, messages)
}

func CallOpenAI(message string, model string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	if apiKey == "" {
		log.Println("OpenAI API key not found")
		return "", errors.New("OpenAI API key not found")
	}

	// 系统提示，包含模型身份
	systemPrompt := fmt.Sprintf("You are %s, a helpful assistant. When asked about your identity or model name, explicitly identify yourself as %s. Please use Markdown format in your responses to make them structured and readable.", model, model)

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": message,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// 构建请求URL，避免/v1路径重复
	var url string
	if baseURL == "" {
		// 如果baseURL为空，使用默认URL
		url = "https://api.openai.com/v1/chat/completions"
	} else if strings.HasSuffix(baseURL, "/v1") {
		// 如果baseURL已经以/v1结尾，直接添加路径
		url = baseURL + "/chat/completions"
	} else {
		// 确保baseURL末尾没有斜杠，然后添加/v1路径
		baseURL = strings.TrimSuffix(baseURL, "/")
		url = baseURL + "/v1/chat/completions"
	}

	log.Printf("Making request to OpenAI API: %s", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
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
		log.Printf("API error (status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v", err)
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	if len(response.Choices) == 0 {
		log.Println("No response from API")
		return "", errors.New("no response from API")
	}

	log.Println("Successfully received response from OpenAI API")
	return response.Choices[0].Message.Content, nil
}

// CallOpenAIStream 使用流式响应调用OpenAI API
func CallOpenAIStream(w http.ResponseWriter, message string, model string) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	if apiKey == "" {
		log.Println("OpenAI API key not found")
		return errors.New("OpenAI API key not found")
	}

	log.Printf("Starting stream request with model: %s", model)

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个有帮助的助手。",
			},
			{
				"role":    "user",
				"content": message,
			},
		},
		"stream": true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return fmt.Errorf("error marshaling request: %v", err)
	}

	// 构建请求URL，避免/v1路径重复
	var url string
	if baseURL == "" {
		url = "https://api.openai.com/v1/chat/completions"
	} else if strings.HasSuffix(baseURL, "/v1") {
		url = baseURL + "/chat/completions"
	} else {
		baseURL = strings.TrimSuffix(baseURL, "/")
		url = baseURL + "/v1/chat/completions"
	}

	log.Printf("Making streaming request to OpenAI API: %s", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{
		Timeout: 120 * time.Second, // 增加超时时间，因为流式响应可能需要更长时间
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to OpenAI: %v", err)
		return fmt.Errorf("error making request to OpenAI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // 允许特定源
	w.Header().Set("X-Accel-Buffering", "no")                              // 禁用Nginx缓冲

	// 发送初始数据确保连接已建立
	fmt.Fprintf(w, "data: \n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	} else {
		log.Println("Warning: ResponseWriter does not support Flush")
	}

	log.Println("Stream connection established, beginning to read response")

	// 创建一个缓冲读取器
	reader := bufio.NewReader(resp.Body)

	// 用于存储完整的响应内容
	var fullContent string

	// 读取流式响应
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("End of OpenAI stream reached")
				break
			}
			log.Printf("Error reading stream from OpenAI: %v", err)
			return fmt.Errorf("error reading stream from OpenAI: %v", err)
		}

		// 去除前缀和空行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "data: [DONE]" {
			log.Println("Received [DONE] signal from OpenAI")
			// 发送完成信号
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			break
		}

		// 解析SSE数据
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var streamResp OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				log.Printf("Error parsing stream data: %v, raw data: %s", err, data)
				continue
			}

			// 提取内容
			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					fullContent += content

					// 发送内容到客户端，确保每个部分都能立即发送
					// 直接发送原始内容，让前端处理格式化
					// 修改：确保内容被正确编码，避免格式问题
					jsonContent, err := json.Marshal(content)
					if err != nil {
						log.Printf("Error marshaling content: %v", err)
						fmt.Fprintf(w, "data: %s\n\n", content)
					} else {
						fmt.Fprintf(w, "data: %s\n\n", string(jsonContent))
					}

					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					} else {
						log.Println("Warning: ResponseWriter does not support Flush")
					}
				}

				// 检查是否完成
				if streamResp.Choices[0].FinishReason != "" {
					log.Printf("Stream finished with reason: %s", streamResp.Choices[0].FinishReason)
				}
			}
		}
	}

	log.Printf("Successfully streamed response from OpenAI API, total length: %d characters", len(fullContent))
	return nil
}

// 添加新的函数，支持传递消息历史
func CallOpenAIStreamWithHistory(w http.ResponseWriter, message string, model string, messages []models.Message) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	if apiKey == "" {
		log.Println("OpenAI API key not found")
		return errors.New("OpenAI API key not found")
	}

	// 记录完整的消息历史以便调试
	log.Printf("Starting stream request with model: %s and %d messages", model, len(messages))
	for i, msg := range messages {
		contentPreview := msg.Content
		if len(contentPreview) > 50 {
			contentPreview = contentPreview[:50] + "..."
		}
		log.Printf("Message %d: Role=%s, Content=%s", i, msg.Role, contentPreview)
	}

	// 转换消息格式
	openaiMessages := make([]map[string]string, 0, len(messages))

	// 如果没有系统提示，添加一个默认的系统提示
	hasSystemPrompt := false
	for _, msg := range messages {
		if msg.Role == "system" {
			hasSystemPrompt = true
			break
		}
	}

	if !hasSystemPrompt {
		// 添加默认的系统提示
		log.Printf("No system prompt found, adding default one with model identity: %s", model)
		systemPrompt := fmt.Sprintf("You are %s, a helpful assistant. When asked about your identity or model name, explicitly identify yourself as %s.", model, model)
		openaiMessages = append(openaiMessages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	// 添加所有消息到请求中
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": openaiMessages,
		"stream":   true,
	}

	// 序列化请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return err
	}

	// 记录请求正文用于调试
	log.Printf("OpenAI request body: %s", string(jsonData))

	// 构建请求
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	// 确保URL格式正确
	url := baseURL
	if !strings.HasSuffix(url, "/v1/chat/completions") {
		url = strings.TrimSuffix(url, "/")
		if strings.HasSuffix(url, "/v1") {
			url += "/chat/completions"
		} else {
			url += "/v1/chat/completions"
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	client := &http.Client{
		Timeout: 180 * time.Second, // 增加超时时间，流式响应可能需要更长时间
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("OpenAI API error: %s, status code: %d", string(body), resp.StatusCode)
		return fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// 读取响应流
	reader := bufio.NewReader(resp.Body)
	fullContent := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading stream: %v", err)
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理特殊情况：[DONE]
		if line == "data: [DONE]" {
			log.Println("Stream complete")
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			break
		}

		// 解析SSE数据
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var streamResp OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				log.Printf("Error parsing stream data: %v, raw data: %s", err, data)
				continue
			}

			// 提取内容
			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					fullContent += content

					// 发送内容到客户端，确保每个部分都能立即发送
					// 直接发送原始内容，让前端处理格式化
					// 修改：确保内容被正确编码，避免格式问题
					jsonContent, err := json.Marshal(content)
					if err != nil {
						log.Printf("Error marshaling content: %v", err)
						fmt.Fprintf(w, "data: %s\n\n", content)
					} else {
						fmt.Fprintf(w, "data: %s\n\n", string(jsonContent))
					}

					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					} else {
						log.Println("Warning: ResponseWriter does not support Flush")
					}
				}

				// 检查是否完成
				if streamResp.Choices[0].FinishReason != "" {
					log.Printf("Stream finished with reason: %s", streamResp.Choices[0].FinishReason)
				}
			}
		}
	}

	log.Printf("Successfully streamed response from OpenAI API, total length: %d characters", len(fullContent))
	return nil
}
