package models

// GetAllValidModels 返回所有支持的模型列表，包括OpenAI和Anthropic模型
func GetAllValidModels() []string {
	var allModels []string

	// 添加OpenAI模型
	openaiModels := GetValidModels()
	allModels = append(allModels, openaiModels...)

	// 添加Anthropic模型
	anthropicModels := []string{
		ModelClaude35Sonnet, // Claude 3.5 Sonnet 2024-10-22
		ModelClaude3Opus,    // Claude 3 Opus
	}
	allModels = append(allModels, anthropicModels...)

	return allModels
}

// GetModelUIName 获取模型的UI显示名称
func GetModelUIName(modelName string) string {
	// 根据模型名称返回用户友好的显示名称
	switch modelName {
	// OpenAI 模型
	case ModelGPT4:
		return "GPT-4"
	case ModelGPT4o:
		return "GPT-4o"
	case ModelGPT4Turbo:
		return "GPT-4 Turbo"
	case ModelGPT35Turbo:
		return "GPT-3.5 Turbo"

	// Anthropic 模型
	case ModelClaude35Sonnet:
		return "Claude 3.5 Sonnet 2024-10-22"
	case ModelClaude3Opus:
		return "Claude 3 Opus"

	default:
		return modelName // 如果没有映射，返回原始名称
	}
}
