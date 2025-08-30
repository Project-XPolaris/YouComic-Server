package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/allentom/haruka"
	"github.com/allentom/harukap/plugins/llm"
	ApiError "github.com/projectxpolaris/youcomic/error"
	"github.com/projectxpolaris/youcomic/plugin"
	"github.com/sirupsen/logrus"
)

// GetLLMConfigHandler 获取当前LLM配置
func GetLLMConfigHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	currentConfig := plugin.LLM.GetCurrentConfig()
	if currentConfig == nil {
		ApiError.RaiseApiError(context, ApiError.LLMConfigNotFoundError, nil)
		return
	}

	// 为了安全，不返回敏感信息如API密钥
	safeConfig := map[string]interface{}{
		"enable":  currentConfig.Enable,
		"default": currentConfig.Default,
	}

	if currentConfig.OpenAI != nil {
		safeConfig["openai"] = map[string]interface{}{
			"enable":   currentConfig.OpenAI.Enable,
			"model":    currentConfig.OpenAI.Model,
			"base_url": currentConfig.OpenAI.BaseURL,
			"has_key":  len(currentConfig.OpenAI.APIKey) > 0,
		}
	}

	if currentConfig.Ollama != nil {
		safeConfig["ollama"] = map[string]interface{}{
			"enable":   currentConfig.Ollama.Enable,
			"base_url": currentConfig.Ollama.BaseURL,
			"model":    currentConfig.Ollama.Model,
		}
	}

	if currentConfig.Gemini != nil {
		safeConfig["gemini"] = map[string]interface{}{
			"enable":   currentConfig.Gemini.Enable,
			"model":    currentConfig.Gemini.Model,
			"location": currentConfig.Gemini.Location,
			"project":  currentConfig.Gemini.Project,
			"has_key":  len(currentConfig.Gemini.APIKey) > 0,
		}
	}

	context.JSONWithStatus(safeConfig, http.StatusOK)
}

// UpdateLLMConfigRequest 更新LLM配置请求
type UpdateLLMConfigRequest struct {
	Enable  bool                       `json:"enable"`
	Default string                     `json:"default"`
	OpenAI  *UpdateOpenAIConfigRequest `json:"openai,omitempty"`
	Ollama  *UpdateOllamaConfigRequest `json:"ollama,omitempty"`
	Gemini  *UpdateGeminiConfigRequest `json:"gemini,omitempty"`
	Persist bool                       `json:"persist"` // 是否持久化到配置文件
}

type UpdateOpenAIConfigRequest struct {
	Enable  bool   `json:"enable"`
	APIKey  string `json:"api_key,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
	Model   string `json:"model"`
}

type UpdateOllamaConfigRequest struct {
	Enable  bool   `json:"enable"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

type UpdateGeminiConfigRequest struct {
	Enable   bool   `json:"enable"`
	APIKey   string `json:"api_key,omitempty"`
	Model    string `json:"model"`
	Location string `json:"location,omitempty"`
	Project  string `json:"project,omitempty"`
}

// UpdateLLMConfigHandler 更新LLM配置
func UpdateLLMConfigHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	var requestBody UpdateLLMConfigRequest
	if err := json.NewDecoder(context.Request.Body).Decode(&requestBody); err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}

	// 获取当前配置
	currentConfig := plugin.LLM.GetCurrentConfig()
	if currentConfig == nil {
		currentConfig = &llm.LLMConfig{}
	}

	// 构建新配置
	newConfig := &llm.LLMConfig{
		Enable:  requestBody.Enable,
		Default: requestBody.Default,
	}

	// 处理OpenAI配置
	if requestBody.OpenAI != nil {
		newConfig.OpenAI = &llm.OpenAIConfig{
			Enable:  requestBody.OpenAI.Enable,
			Model:   requestBody.OpenAI.Model,
			BaseURL: requestBody.OpenAI.BaseURL,
		}
		// 如果没有提供新的API Key，使用现有的
		if requestBody.OpenAI.APIKey != "" {
			newConfig.OpenAI.APIKey = requestBody.OpenAI.APIKey
		} else if currentConfig.OpenAI != nil {
			newConfig.OpenAI.APIKey = currentConfig.OpenAI.APIKey
		}
	}

	// 处理Ollama配置
	if requestBody.Ollama != nil {
		newConfig.Ollama = &llm.OllamaConfig{
			Enable:  requestBody.Ollama.Enable,
			BaseURL: requestBody.Ollama.BaseURL,
			Model:   requestBody.Ollama.Model,
		}
	}

	// 处理Gemini配置
	if requestBody.Gemini != nil {
		newConfig.Gemini = &llm.GeminiConfig{
			Enable:   requestBody.Gemini.Enable,
			Model:    requestBody.Gemini.Model,
			Location: requestBody.Gemini.Location,
			Project:  requestBody.Gemini.Project,
		}
		// 如果没有提供新的API Key，使用现有的
		if requestBody.Gemini.APIKey != "" {
			newConfig.Gemini.APIKey = requestBody.Gemini.APIKey
		} else if currentConfig.Gemini != nil {
			newConfig.Gemini.APIKey = currentConfig.Gemini.APIKey
		}
	}

	// 更新配置
	var err error
	if requestBody.Persist {
		err = plugin.LLM.UpdateAndSaveConfig(newConfig)
	} else {
		err = plugin.LLM.UpdateConfig(newConfig)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to update LLM config")
		ApiError.RaiseApiError(context, ApiError.LLMConfigInvalidError, nil)
		return
	}

	// 返回更新后的配置（不包含敏感信息）
	GetLLMConfigHandler(context)
}

// ReloadLLMConfigHandler 从配置文件重新加载LLM配置
func ReloadLLMConfigHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	err := plugin.LLM.ReloadConfig()
	if err != nil {
		logrus.WithError(err).Error("Failed to reload LLM config")
		ApiError.RaiseApiError(context, ApiError.LLMConfigInvalidError, nil)
		return
	}

	// 返回重新加载后的配置
	GetLLMConfigHandler(context)
}

// SaveLLMConfigHandler 保存当前LLM配置到配置文件
func SaveLLMConfigHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	err := plugin.LLM.SaveConfig()
	if err != nil {
		logrus.WithError(err).Error("Failed to save LLM config")
		context.JSONWithStatus(map[string]interface{}{
			"success": false,
			"message": "Configuration save failed",
			"error":   err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	context.JSONWithStatus(map[string]interface{}{
		"success": true,
		"message": "Configuration saved successfully",
	}, http.StatusOK)
}

// GetLLMStatusHandler 获取LLM插件状态
func GetLLMStatusHandler(context *haruka.Context) {
	status := map[string]interface{}{
		"available": false,
		"providers": []string{},
		"default":   "",
	}

	if plugin.LLM != nil {
		status["available"] = true
		status["providers"] = plugin.LLM.GetAvailableProviders()
		status["default"] = plugin.LLM.GetDefaultProvider()
	}

	context.JSONWithStatus(status, http.StatusOK)
}

// TestLLMConnectionRequest 测试LLM连接请求
type TestLLMConnectionRequest struct {
	Provider string `json:"provider,omitempty"` // 可选，测试特定提供商
	Prompt   string `json:"prompt,omitempty"`   // 可选，自定义测试提示
}

// TestLLMConnectionHandler 测试LLM连接
func TestLLMConnectionHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	var requestBody TestLLMConnectionRequest
	if err := json.NewDecoder(context.Request.Body).Decode(&requestBody); err != nil {
		// 如果没有请求体，使用默认值
		requestBody = TestLLMConnectionRequest{}
	}

	// 设置默认测试提示
	prompt := requestBody.Prompt
	if prompt == "" {
		prompt = "Hello, this is a test message. Please respond with 'Test successful'."
	}

	var response string
	var err error

	// 如果指定了提供商，测试特定提供商
	if requestBody.Provider != "" {
		response, err = plugin.LLM.GenerateTextWithProvider(context.Request.Context(), prompt, requestBody.Provider)
	} else {
		// 测试默认提供商
		response, err = plugin.LLM.GenerateText(context.Request.Context(), prompt)
	}

	result := map[string]interface{}{
		"success": err == nil,
		"prompt":  prompt,
	}

	if err != nil {
		result["error"] = err.Error()
		logrus.WithError(err).Warn("LLM connection test failed")
	} else {
		result["response"] = response
		result["response_length"] = len(response)
	}

	context.JSONWithStatus(result, http.StatusOK)
}
