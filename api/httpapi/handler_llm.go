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

	// 添加模板配置
	if currentConfig.TemplateConfig != nil {
		templateConfig := map[string]interface{}{
			"default_scenario": currentConfig.TemplateConfig.DefaultScenario,
		}

		// 构建业务场景配置
		if currentConfig.TemplateConfig.BusinessScenarios != nil {
			scenarios := make(map[string]interface{})
			for key, scenario := range currentConfig.TemplateConfig.BusinessScenarios {
				scenarios[key] = map[string]interface{}{
					"name":             scenario.Name,
					"description":      scenario.Description,
					"default_template": scenario.DefaultTemplate,
					"custom_templates": scenario.CustomTemplates,
					"active_template":  scenario.ActiveTemplate,
					"variables":        scenario.Variables,
				}
			}
			templateConfig["business_scenarios"] = scenarios
		}

		safeConfig["template_config"] = templateConfig
	}

	context.JSONWithStatus(safeConfig, http.StatusOK)
}

// RenderTemplateRequest 渲染模板请求
type RenderTemplateRequest struct {
	Scenario  string            `json:"scenario"`  // 业务场景名称
	Variables map[string]string `json:"variables"` // 模板变量
}

// RenderTemplateResponse 渲染模板响应
type RenderTemplateResponse struct {
	RenderedText string `json:"rendered_text"` // 渲染后的文本
	Scenario     string `json:"scenario"`      // 业务场景名称
}

// RenderTemplateHandler 渲染模板
func RenderTemplateHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	var requestBody RenderTemplateRequest
	if err := json.NewDecoder(context.Request.Body).Decode(&requestBody); err != nil {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}

	if requestBody.Scenario == "" {
		ApiError.RaiseApiError(context, ApiError.JsonParseError, nil)
		return
	}

	// 使用LLM插件渲染模板
	renderedText, err := plugin.LLM.RenderTemplate(requestBody.Scenario, requestBody.Variables)
	if err != nil {
		logrus.WithError(err).Error("Failed to render template")
		ApiError.RaiseApiError(context, ApiError.LLMConfigInvalidError, nil)
		return
	}

	response := RenderTemplateResponse{
		RenderedText: renderedText,
		Scenario:     requestBody.Scenario,
	}

	context.JSONWithStatus(response, http.StatusOK)
}

// GetScenariosHandler 获取所有可用的业务场景
func GetScenariosHandler(context *haruka.Context) {
	if plugin.LLM == nil {
		ApiError.RaiseApiError(context, ApiError.LLMPluginNotAvailableError, nil)
		return
	}

	scenarios := plugin.LLM.GetAvailableScenarios()

	// 获取每个场景的详细信息
	scenarioInfos := make(map[string]interface{})
	for _, scenarioName := range scenarios {
		if scenarioInfo, err := plugin.LLM.GetScenarioInfo(scenarioName); err == nil {
			scenarioInfos[scenarioName] = map[string]interface{}{
				"name":        scenarioInfo.Name,
				"description": scenarioInfo.Description,
				"variables":   scenarioInfo.Variables,
			}
		}
	}

	response := map[string]interface{}{
		"scenarios": scenarios,
		"details":   scenarioInfos,
	}

	context.JSONWithStatus(response, http.StatusOK)
}

// UpdateLLMConfigRequest 更新LLM配置请求
type UpdateLLMConfigRequest struct {
	Enable         bool                         `json:"enable"`
	Default        string                       `json:"default"`
	OpenAI         *UpdateOpenAIConfigRequest   `json:"openai,omitempty"`
	Ollama         *UpdateOllamaConfigRequest   `json:"ollama,omitempty"`
	Gemini         *UpdateGeminiConfigRequest   `json:"gemini,omitempty"`
	TemplateConfig *UpdateTemplateConfigRequest `json:"template_config,omitempty"`
	Persist        bool                         `json:"persist"` // 是否持久化到配置文件
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

type UpdateTemplateConfigRequest struct {
	DefaultScenario   string                                    `json:"default_scenario,omitempty"`
	BusinessScenarios map[string]*UpdateBusinessScenarioRequest `json:"business_scenarios,omitempty"`
}

type UpdateBusinessScenarioRequest struct {
	// 注意：Name、Description、Variables字段被移除，因为这些是预定义的
	DefaultTemplate string            `json:"default_template,omitempty"`
	CustomTemplates map[string]string `json:"custom_templates,omitempty"`
	ActiveTemplate  string            `json:"active_template,omitempty"`
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

	// 处理模板配置
	if requestBody.TemplateConfig != nil {
		newConfig.TemplateConfig = &llm.TemplateConfig{
			DefaultScenario:   requestBody.TemplateConfig.DefaultScenario,
			BusinessScenarios: make(map[string]*llm.BusinessScenarioConfig),
		}

		// 处理业务场景配置 - 只更新用户可配置的部分
		if requestBody.TemplateConfig.BusinessScenarios != nil {
			// 先确保有预定义的场景配置
			if newConfig.TemplateConfig.BusinessScenarios == nil {
				newConfig.TemplateConfig.BusinessScenarios = make(map[string]*llm.BusinessScenarioConfig)
			}

			// 加载预定义场景作为基础
			predefinedScenarios := llm.GetPredefinedScenarios()
			for key, predefined := range predefinedScenarios {
				if _, exists := newConfig.TemplateConfig.BusinessScenarios[key]; !exists {
					// 复制预定义场景
					newConfig.TemplateConfig.BusinessScenarios[key] = &llm.BusinessScenarioConfig{
						Name:            predefined.Name,
						Description:     predefined.Description,
						DefaultTemplate: predefined.DefaultTemplate,
						CustomTemplates: make(map[string]string),
						ActiveTemplate:  predefined.ActiveTemplate,
						Variables:       predefined.Variables,
					}
				}
			}

			// 只更新用户可配置的字段
			for key, scenarioReq := range requestBody.TemplateConfig.BusinessScenarios {
				if existingScenario, exists := newConfig.TemplateConfig.BusinessScenarios[key]; exists {
					// 只更新用户可配置的字段，保持预定义字段不变
					if scenarioReq.DefaultTemplate != "" {
						existingScenario.DefaultTemplate = scenarioReq.DefaultTemplate
					}
					if scenarioReq.CustomTemplates != nil {
						existingScenario.CustomTemplates = scenarioReq.CustomTemplates
					}
					if scenarioReq.ActiveTemplate != "" || (scenarioReq.ActiveTemplate == "" && scenarioReq.CustomTemplates != nil) {
						existingScenario.ActiveTemplate = scenarioReq.ActiveTemplate
					}
				}
			}
		}
	} else if currentConfig.TemplateConfig != nil {
		// 保留现有的模板配置
		newConfig.TemplateConfig = currentConfig.TemplateConfig
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

// getDefaultTagPrompt 获取默认的标签分析prompt
func getDefaultTagPrompt() string {
	return `你是一个专业的漫画标签分析助手。请分析以下文本，提取出漫画相关的标签信息。

文本内容："%s"

请从文本中提取以下类型的标签：
- artist: 画师/作者名称
- series: 系列/作品名称  
- name: 漫画标题/名称
- theme: 主题/题材标签
- translator: 翻译者
- type: 漫画类型(如CG、同人志等)
- lang: 语言
- magazine: 杂志名称
- societies: 社团名称

请以JSON格式返回结果，格式如下：
{
  "tags": [
    {"name": "标签名称", "type": "标签类型"},
    {"name": "标签名称", "type": "标签类型"}
  ]
}

要求：
1. 只提取确实存在于文本中的信息
2. 标签名称要准确，去除多余符号
3. 如果无法确定标签类型，可以留空
4. 不要添加文本中不存在的信息
5. 返回纯JSON，不要其他说明文字`
}
