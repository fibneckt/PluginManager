package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"PluginsManager/model"
)

func init() {
	p := &AIPlugin{
		LLMURL:    "https://api.deepseek.com/chat/completions",
		LLMAPIKey: "your-api-key-here",
		Model:     "deepseek-chat",
		MaxTokens: 2048,
	}
	if err := model.RegisterPlugin(p.Name(), p.Version(), p); err != nil {
		panic("register ai: " + err.Error())
	}
}

// AIPlugin AI 对话插件
type AIPlugin struct {
	LLMURL    string
	LLMAPIKey string
	Model     string
	MaxTokens int
}

func (p *AIPlugin) Name() string    { return "ai" }
func (p *AIPlugin) Version() string { return "1.0.0" }

func (p *AIPlugin) Configure(config map[string]interface{}) error {
	if v, ok := config["llm_url"].(string); ok {
		p.LLMURL = v
	}
	if v, ok := config["llm_api_key"].(string); ok {
		p.LLMAPIKey = v
	}
	if v, ok := config["model"].(string); ok {
		p.Model = v
	}
	if v, ok := config["max_tokens"]; ok {
		switch val := v.(type) {
		case int:
			p.MaxTokens = val
		case float64:
			p.MaxTokens = int(val)
		}
	}
	return nil
}

// Run 发送对话到 LLM 并返回回复
// 输入: {"messages": [{"role":"user","content":"你好"}]}
// 输出: {"response": "..."}
func (p *AIPlugin) Run(data map[string]interface{}) (map[string]interface{}, error) {
	messages, ok := data["messages"]
	if !ok {
		return nil, fmt.Errorf("missing messages field")
	}

	body := map[string]interface{}{
		"model":       p.Model,
		"messages":    messages,
		"max_tokens":  p.MaxTokens,
		"temperature": 0.7,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", p.LLMURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.LLMAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %v", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return map[string]interface{}{
		"response": result.Choices[0].Message.Content,
	}, nil
}
