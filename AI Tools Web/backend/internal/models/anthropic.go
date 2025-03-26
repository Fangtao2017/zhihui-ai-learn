package models

// 定义 Anthropic API 相关的结构体
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
}

// 添加 Anthropic 可用模型常量
const (
	// Claude 模型系列 - 只保留最常用的两个模型
	ModelClaude35Sonnet = "claude-3-5-sonnet-20241022" // 最新版本: 2024-10-22
	ModelClaude3Opus    = "claude-3-opus-20240229"     // Opus模型
)

// 模型别名映射，支持前端简化的模型名称
var ModelAliases = map[string]string{
	"claude-3-5-sonnet":            ModelClaude35Sonnet, // 默认使用最新版本
	"claude-3-5-sonnet-2024-10-22": ModelClaude35Sonnet, // 支持带日期的版本
	"Claude 3.5 Sonnet 2024-10-22": ModelClaude35Sonnet, // 完整UI显示名称
	"claude-3-opus":                ModelClaude3Opus,    // Opus模型
	"Claude 3 Opus":                ModelClaude3Opus,    // 完整UI显示名称
}

// IsAnthropicModel checks if a model is an Anthropic model
func IsAnthropicModel(model string) bool {
	// 首先检查模型别名
	if _, ok := ModelAliases[model]; ok {
		return true
	}

	// Check if model starts with "claude" (case insensitive)
	if len(model) >= 6 && (model[:6] == "claude" || model[:6] == "Claude") {
		return true
	}

	// Check against known model constants
	switch model {
	case ModelClaude35Sonnet, ModelClaude3Opus:
		return true
	}

	return false
}
