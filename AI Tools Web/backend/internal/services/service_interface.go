package services

import (
	"backend/internal/models"
	"net/http"
)

// LLMService defines the interface for language model services
type LLMService interface {
	// GetModelName returns the name of the model
	GetModelName() string

	// GetModelProvider returns the provider of the model (OpenAI, Anthropic, etc.)
	GetModelProvider() string

	// CallModel calls the model with a single message and returns the response
	CallModel(message string, model string) (string, error)

	// CallModelStreamWithHistory calls the model with a stream response and message history
	CallModelStreamWithHistory(w http.ResponseWriter, message string, model string, messages []models.Message) error
}

// GetModelProvider returns the provider of the model
func GetModelProvider(model string) string {
	// 先检查简化的模型名称是否有别名
	if _, ok := models.ModelAliases[model]; ok {
		return "anthropic"
	}

	// Check if model is an Anthropic model
	if models.IsAnthropicModel(model) {
		return "anthropic"
	}

	// Default to OpenAI
	return "openai"
}

// GetLLMService returns the appropriate service for the model
func GetLLMService(model string) LLMService {
	provider := GetModelProvider(model)

	switch provider {
	case "anthropic":
		return &AnthropicService{}
	default:
		return &OpenAIService{}
	}
}
