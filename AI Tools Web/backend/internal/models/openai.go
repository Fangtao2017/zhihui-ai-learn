package models

type OpenAIConfig struct {
	APIKey string
	Model  string
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// 添加 OpenAI 可用模型常量
const (
	ModelGPT4       = "gpt-4"  // 修正为正确的GPT-4
	ModelGPT4o      = "gpt-4o" // 添加真正的GPT-4o模型
	ModelGPT4Turbo  = "gpt-4-turbo"
	ModelGPT35Turbo = "gpt-3.5-turbo"
)

// 定义模型所属的提供商
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
)

// 获取有效的模型列表
func GetValidModels() []string {
	var validModels []string

	// 只添加 OpenAI 模型
	validModels = append(validModels, []string{
		ModelGPT4,
		ModelGPT4o,
		ModelGPT4Turbo,
		ModelGPT35Turbo,
	}...)

	// 不再在OpenAI文件中引用Anthropic模型
	// 相关代码已移至anthropic.go

	return validModels
}

// 根据模型名称获取对应的提供商
func GetModelProvider(model string) string {
	switch model {
	case ModelGPT4, ModelGPT4o, ModelGPT4Turbo, ModelGPT35Turbo:
		return ProviderOpenAI
	default:
		// 不在这里处理Claude模型，交给anthropic.go中的函数处理
		// 如果模型名不是OpenAI模型，应该先检查是否是Anthropic模型
		if IsAnthropicModel(model) {
			return ProviderAnthropic
		}
		return ProviderOpenAI // 默认使用OpenAI
	}
}
