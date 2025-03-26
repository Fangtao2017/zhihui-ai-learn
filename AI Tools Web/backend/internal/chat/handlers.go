package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/services"

	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	repo := db.NewChatRepository()
	history, err := repo.GetChatHistory(r.Context())
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value("user").(auth.UserClaims)
	chatID := primitive.NewObjectID()

	log.Printf("UserClaims: %v", userClaims)

	// 解析请求体，允许客户端指定模型
	var req struct {
		Title string `json:"title"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 如果请求体解析失败，使用默认值
		req.Title = "New Chat"
		req.Model = "gpt-3.5-turbo" // 默认模型
	}

	// 如果没有提供标题，使用默认值
	if req.Title == "" {
		req.Title = "New Chat"
	}

	// 如果没有提供模型或模型无效，使用默认值
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	} else {
		// 验证模型是否有效
		validModels := models.GetAllValidModels()
		isValidModel := false
		for _, model := range validModels {
			if model == req.Model {
				isValidModel = true
				break
			}
		}

		if !isValidModel {
			log.Printf("Invalid model specified: %s, using default model", req.Model)
			req.Model = "gpt-3.5-turbo"
		}
	}

	chat := &models.Chat{
		ID:        chatID.Hex(),
		UserID:    userClaims.Email,
		Title:     req.Title,
		Model:     req.Model,
		CreatedAt: time.Now(),
	}

	repo := db.NewChatRepository()
	if err := repo.CreateChat(r.Context(), chat); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["id"]

	repo := db.NewChatRepository()
	messages, err := repo.GetMessages(r.Context(), chatID)
	if err != nil {
		// 检查错误类型，如果是"没有找到消息"类型的错误，返回空数组而不是错误
		if strings.Contains(err.Error(), "no documents") || strings.Contains(err.Error(), "not found") {
			log.Printf("Chat %s exists but has no messages, returning empty array", chatID)
			w.Header().Set("Content-Type", "application/json")
			// 明确返回空数组，不要返回null
			json.NewEncoder(w).Encode([]models.Message{})
			return
		}

		log.Printf("Error getting messages: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	// 确保返回的是一个有效的JSON数组，即使messages为空
	w.Header().Set("Content-Type", "application/json")
	if messages == nil {
		// 如果messages为nil，明确返回空数组
		json.NewEncoder(w).Encode([]models.Message{})
	} else {
		json.NewEncoder(w).Encode(messages)
	}
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["id"]

	// 验证 chatID 是否有效
	if chatID == "" {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Message string `json:"message"`
		Model   string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 获取聊天仓库实例
	repo := db.NewChatRepository()

	// 获取聊天信息
	chatInfo, err := repo.GetChat(r.Context(), chatID)
	if err != nil {
		log.Printf("Warning: Failed to get chat info: %v. Creating new record.", err)
		// 如果聊天记录不存在，尝试创建一个新的
		chatInfo = &models.Chat{
			ID:        chatID,
			Title:     "New Chat",
			Model:     req.Model,
			CreatedAt: time.Now(),
		}
		if err := repo.CreateChat(r.Context(), chatInfo); err != nil {
			log.Printf("Error creating chat: %v", err)
			http.Error(w, "Failed to create chat", http.StatusInternalServerError)
			return
		}
	} else if chatInfo.Model == "" && req.Model != "" {
		// 如果聊天存在但模型字段为空，使用请求中的模型
		if err := repo.UpdateChatModel(r.Context(), chatID, req.Model); err != nil {
			log.Printf("Error updating chat model: %v", err)
			// 继续执行，不中断处理
		} else {
			chatInfo.Model = req.Model
		}
	}

	// 确定使用的模型: 优先使用聊天记录中的模型，其次是请求中的模型，最后是默认模型
	model := chatInfo.Model
	log.Printf("Initial model from chat history: %s", model)

	if model == "" && req.Model != "" {
		model = req.Model
		log.Printf("Using model from request: %s", model)
	}

	if model == "" {
		model = "gpt-3.5-turbo"
		log.Printf("Using default model: %s", model)
	}

	log.Printf("Final selected model: %s", model)
	log.Printf("Is Anthropic model? %v", models.IsAnthropicModel(model))

	// 检查是否有别名映射
	if mappedModel, ok := models.ModelAliases[model]; ok {
		log.Printf("Model has alias mapping: %s -> %s", model, mappedModel)
	}

	// 保存用户消息
	userMessage := &models.Message{
		ChatID:  chatID,
		Role:    "user",
		Content: req.Message,
	}

	if err := repo.SaveMessage(r.Context(), userMessage); err != nil {
		log.Printf("Error saving user message: %v", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	// 使用服务接口调用相应的LLM服务
	llmService := services.GetLLMService(model)
	log.Printf("Using LLM service: %s for model: %s", llmService.GetModelProvider(), model)

	aiResponse, apiErr := llmService.CallModel(req.Message, model)
	if apiErr != nil {
		log.Printf("Error calling AI API: %v", apiErr)

		// 检查是否是Claude模型错误，如果是，尝试回退到OpenAI
		if models.IsAnthropicModel(model) && (strings.Contains(apiErr.Error(), "API key not found") ||
			strings.Contains(apiErr.Error(), "Anthropic API") ||
			strings.Contains(apiErr.Error(), "No content received")) {
			// 回退到使用OpenAI模型
			fallbackModel := "gpt-3.5-turbo"
			log.Printf("Falling back to %s due to Anthropic API error", fallbackModel)

			// 使用OpenAI服务
			openaiService := &services.OpenAIService{}
			fallbackResponse, fallbackErr := openaiService.CallModel(req.Message, fallbackModel)

			if fallbackErr != nil {
				log.Printf("Error calling fallback model: %v", fallbackErr)
				http.Error(w, "Failed to get AI response from both primary and fallback models", http.StatusInternalServerError)
				return
			}

			// 更新聊天记录中的模型
			if err := repo.UpdateChatModel(r.Context(), chatID, fallbackModel); err != nil {
				log.Printf("Error updating chat model to fallback: %v", err)
			}

			log.Printf("Successfully used fallback model for chat ID: %s", chatID)
			aiResponse = fallbackResponse
		} else {
			// 其他错误直接返回
			http.Error(w, "Failed to get AI response", http.StatusInternalServerError)
			return
		}
	}

	// 保存 AI 回复
	aiMessage := &models.Message{
		ChatID:  chatID,
		Role:    "assistant",
		Content: aiResponse,
	}

	if err := repo.SaveMessage(r.Context(), aiMessage); err != nil {
		log.Printf("Error saving AI message: %v", err)
		http.Error(w, "Failed to save AI response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response": aiResponse,
	})
}

func UpdateChatTitleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["id"]

	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	repo := db.NewChatRepository()
	if err := repo.UpdateChatTitle(r.Context(), chatID, req.Title); err != nil {
		log.Printf("Error updating chat title: %v", err)
		http.Error(w, "Failed to update chat title", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Title updated successfully",
	})
}

// SendMessageStreamHandler 处理流式发送消息的请求
func SendMessageStreamHandler(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取聊天ID
	vars := mux.Vars(r)
	chatID := vars["id"]
	if chatID == "" {
		log.Println("Missing chat ID")
		http.Error(w, "Missing chat ID", http.StatusBadRequest)
		return
	}

	// 获取聊天信息
	repo := db.NewChatRepository()
	chatInfo, err := repo.GetChat(r.Context(), chatID)
	if err != nil {
		log.Printf("Error getting chat info: %v", err)

		// 如果找不到聊天记录，创建一个默认的记录
		if strings.Contains(err.Error(), "no documents") || strings.Contains(err.Error(), "not found") {
			log.Printf("Creating temporary chat info for streaming response, chat ID: %s", chatID)
			chatInfo = &models.Chat{
				ID:        chatID,
				Title:     "New Chat",
				CreatedAt: time.Now(),
			}
		} else {
			// 如果是其他错误，返回错误信息
			fmt.Fprintf(w, "data: ERROR: Failed to get chat information\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}
	}

	// 获取聊天历史
	messages, err := repo.GetMessages(r.Context(), chatID)
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
		// 如果是"没有找到消息"类型的错误，设置为空数组
		if strings.Contains(err.Error(), "no documents") || strings.Contains(err.Error(), "not found") {
			log.Printf("No messages found for chat %s, using empty history", chatID)
			messages = []models.Message{}
		} else {
			// 其他错误也设置为空数组，允许继续流式响应
			messages = []models.Message{}
		}
	}

	// 获取消息内容
	message := r.URL.Query().Get("message")

	// 使用聊天保存的模型，如果没有则使用查询参数中的模型作为备用
	model := chatInfo.Model
	if model == "" {
		model = r.URL.Query().Get("model")
		// 如果模型仍然为空，使用默认模型
		if model == "" {
			model = "gpt-3.5-turbo"
		}

		// 更新聊天的模型
		err := repo.UpdateChatModel(r.Context(), chatID, model)
		if err != nil {
			log.Printf("Error updating chat model: %v", err)
			// 继续执行，不中断处理
		}
	}

	// 获取语言偏好
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "english" // Default to English, consistent with frontend
	}

	// Try to detect user input language (if frontend specifies auto or not specified)
	if language == "auto" {
		// Simple language detection: calculate the ratio of Chinese characters
		// Chinese character range is roughly: \u4e00-\u9fff
		totalChars := 0
		chineseChars := 0

		for _, r := range message {
			if r > ' ' { // Ignore whitespace
				totalChars++
				if r >= 0x4e00 && r <= 0x9fff {
					chineseChars++
				}
			}
		}

		// If Chinese characters are more than 15%, consider it Chinese input
		if totalChars > 0 && float64(chineseChars)/float64(totalChars) > 0.15 {
			language = "chinese"
		} else {
			language = "english"
		}

		log.Printf("Language auto-detection: total chars=%d, Chinese chars=%d, detected language=%s",
			totalChars, chineseChars, language)
	}

	log.Printf("Received stream request for chat ID: %s, model: %s, language: %s", chatID, model, language)

	// 保存用户消息到数据库
	userMessage := &models.Message{
		ChatID:    chatID,
		Role:      "user",
		Content:   message,
		CreatedAt: time.Now(),
	}

	if err := repo.SaveMessage(r.Context(), userMessage); err != nil {
		log.Printf("Error saving user message: %v", err)
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Build system prompt based on language preference and model information
	systemPrompt := ""
	if language == "english" {
		systemPrompt = fmt.Sprintf("You are %s, a helpful assistant. When asked about your identity or model name, explicitly identify yourself as %s. Please respond in English.", model, model)
	} else {
		systemPrompt = fmt.Sprintf("You are %s, a helpful assistant. When asked about your identity or model name, explicitly identify yourself as %s. Please respond in Chinese.", model, model)
	}

	// 添加调试日志，确认模型和系统提示
	log.Printf("Sending request with model: %s", model)
	log.Printf("System prompt: %s", systemPrompt)

	// Build complete message history, including system prompt
	var fullMessages []models.Message
	// Add system prompt
	fullMessages = append(fullMessages, models.Message{
		Role:    "system",
		Content: systemPrompt,
	})
	// Add historical messages
	fullMessages = append(fullMessages, messages...)
	// Add current user message
	fullMessages = append(fullMessages, *userMessage)

	// 使用新的服务接口
	llmService := services.GetLLMService(model)
	log.Printf("Using LLM service: %s for model: %s", llmService.GetModelProvider(), model)

	apiErr := llmService.CallModelStreamWithHistory(w, message, model, fullMessages)
	if apiErr != nil {
		log.Printf("Error calling AI stream: %v", apiErr)

		// 检查是否是Claude模型错误，如果是，尝试回退到OpenAI
		if models.IsAnthropicModel(model) && (strings.Contains(apiErr.Error(), "API key not found") ||
			strings.Contains(apiErr.Error(), "Anthropic API") ||
			strings.Contains(apiErr.Error(), "No content received")) {
			// 回退到使用OpenAI模型
			fallbackModel := "gpt-3.5-turbo"
			log.Printf("Falling back to %s due to Anthropic API error", fallbackModel)

			// 通知客户端
			fmt.Fprintf(w, "data: ERROR: Claude模型不可用，正在使用%s代替\n\n", fallbackModel)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			// 更新系统提示信息
			for i, msg := range fullMessages {
				if msg.Role == "system" {
					if language == "english" {
						fullMessages[i].Content = fmt.Sprintf("You are %s, a helpful assistant. When asked about your identity or model name, explicitly identify yourself as %s. Please respond in English.", fallbackModel, fallbackModel)
					} else {
						fullMessages[i].Content = fmt.Sprintf("You are %s, a helpful assistant. When asked about your identity or model name, explicitly identify yourself as %s. Please respond in Chinese.", fallbackModel, fallbackModel)
					}
					break
				}
			}

			// 使用OpenAI服务
			openaiService := &services.OpenAIService{}
			fallbackErr := openaiService.CallModelStreamWithHistory(w, message, fallbackModel, fullMessages)

			if fallbackErr != nil {
				log.Printf("Error calling fallback model: %v", fallbackErr)
				fmt.Fprintf(w, "data: ERROR: %s\n\n", fallbackErr.Error())
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				return
			}

			// 更新聊天记录中的模型
			if err := repo.UpdateChatModel(r.Context(), chatID, fallbackModel); err != nil {
				log.Printf("Error updating chat model to fallback: %v", err)
			}

			log.Printf("Successfully used fallback model for chat ID: %s", chatID)
			return
		}

		// 其他错误直接返回
		fmt.Fprintf(w, "data: ERROR: %s\n\n", apiErr.Error())
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	log.Printf("Stream completed for chat ID: %s", chatID)
}

// SaveAIMessageHandler 保存AI回复
func SaveAIMessageHandler(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	chatID := vars["id"]

	// 验证 chatID 是否有效
	if chatID == "" {
		log.Printf("Invalid chat ID in SaveAIMessageHandler")
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// 记录请求体用于调试
	if len(body) > 100 {
		log.Printf("SaveAIMessageHandler: Received AI message for chat %s, content length: %d, preview: %s...",
			chatID, len(body), string(body)[:100])
	} else if len(body) > 0 {
		log.Printf("SaveAIMessageHandler: Received AI message for chat %s, content length: %d, full content: %s",
			chatID, len(body), string(body))
	} else {
		log.Printf("SaveAIMessageHandler: Received empty AI message for chat %s", chatID)
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		log.Printf("Empty content in SaveAIMessageHandler for chat %s", chatID)
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	// 保存 AI 回复
	aiMessage := &models.Message{
		ChatID:    chatID,
		Role:      "assistant",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	repo := db.NewChatRepository()
	if err := repo.SaveMessage(r.Context(), aiMessage); err != nil {
		log.Printf("Error saving AI message to chat %s: %v", chatID, err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved AI message to chat %s, content length: %d", chatID, len(req.Content))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// DeleteChatHandler 删除聊天
func DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		http.Error(w, "Missing chat ID", http.StatusBadRequest)
		return
	}

	repo := db.NewChatRepository()
	if err := repo.DeleteChat(r.Context(), chatID); err != nil {
		log.Printf("Error deleting chat: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Chat deleted successfully"})
}

// GetChatInfoHandler 获取聊天信息
func GetChatInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		http.Error(w, "Missing chat ID", http.StatusBadRequest)
		return
	}

	repo := db.NewChatRepository()
	chat, err := repo.GetChat(r.Context(), chatID)
	if err != nil {
		// 如果是找不到聊天的错误，可能是数据库错误或真的不存在
		log.Printf("Error getting chat info: %v", err)

		// 尝试创建一个默认的聊天信息返回
		if strings.Contains(err.Error(), "no documents") || strings.Contains(err.Error(), "not found") {
			log.Printf("Creating default chat info for missing chat ID: %s", chatID)
			defaultChat := &models.Chat{
				ID:        chatID,
				Title:     "New Chat",
				CreatedAt: time.Now(),
			}

			// 不要试图保存这个默认聊天，只是返回给客户端
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(defaultChat)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

// UpdateChatModelHandler 更新聊天使用的模型
func UpdateChatModelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["id"]

	if chatID == "" {
		http.Error(w, "Missing chat ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Model string `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证模型是否有效
	validModels := models.GetAllValidModels()
	isValidModel := false
	for _, model := range validModels {
		if model == req.Model {
			isValidModel = true
			break
		}
	}

	if !isValidModel {
		log.Printf("Invalid model: %s", req.Model)
		http.Error(w, "Invalid model", http.StatusBadRequest)
		return
	}

	repo := db.NewChatRepository()
	if err := repo.UpdateChatModel(r.Context(), chatID, req.Model); err != nil {
		log.Printf("Error updating chat model: %v", err)
		http.Error(w, "Failed to update chat model", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Model updated successfully",
	})
}

// GetAvailableModelsHandler 返回所有可用的模型列表
func GetAvailableModelsHandler(w http.ResponseWriter, r *http.Request) {
	// 获取所有支持的模型
	availableModels := models.GetAllValidModels()

	// 构建响应数据，包含模型ID和用户友好的显示名称
	var modelsList []map[string]string
	for _, modelID := range availableModels {
		displayName := models.GetModelUIName(modelID)
		provider := models.GetModelProvider(modelID)

		modelsList = append(modelsList, map[string]string{
			"id":       modelID,
			"name":     displayName,
			"provider": provider,
		})
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(modelsList); err != nil {
		log.Printf("Error encoding models list: %v", err)
		http.Error(w, "Failed to encode models list", http.StatusInternalServerError)
		return
	}
}
