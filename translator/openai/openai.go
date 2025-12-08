package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	oai "github.com/sashabaranov/go-openai"
	"github.com/smilingpoplar/translate/config"
	"github.com/smilingpoplar/translate/translator/middleware"
	"github.com/smilingpoplar/translate/translator/transerrors"
	"github.com/smilingpoplar/translate/util"
)

type OpenAI struct {
	config  *oai.ClientConfig
	client  *oai.Client
	model   string
	prompt  string
	handler middleware.Handler
	onTrans func([]string) error
	Name    string
	reqArgs map[string]any
	// apiKey is stored only for GLM thinking requests since OpenAI library doesn't support this parameter
	apiKey string
}

type option func(*OpenAI) error

func New(name, key, baseURL, model string, opts ...option) (*OpenAI, error) {
	o := &OpenAI{Name: name, model: model}
	chain := middleware.Chain(
		middleware.TextsLimit(3000),
		middleware.OnTranslated(&o.onTrans),
		middleware.RetryWithCache(name, 3, 8),
	)
	o.handler = chain(o.translate)

	config := oai.DefaultConfig(key)
	config.BaseURL = baseURL
	o.config = &config
	// Store API key for GLM custom requests only when needed
	o.apiKey = key
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, fmt.Errorf("error creating openai translator: %w", err)
		}
	}
	o.client = oai.NewClientWithConfig(config)

	return o, nil
}

func WithProxy(proxy string) option {
	return func(o *OpenAI) error {
		return util.SetProxy(proxy, o.config.HTTPClient.(*http.Client))
	}
}

func WithReqArgs(req map[string]any) option {
	return func(o *OpenAI) error {
		o.reqArgs = req
		return nil
	}
}

func (o *OpenAI) translate(texts []string, toLang string) ([]string, error) {
	prompt := o.prompt
	if prompt == "" {
		var err error
		prompt, err = config.GetPrompt(texts, toLang)
		if err != nil {
			return nil, fmt.Errorf("error translating: %w", err)
		}
	}

	result, err := o.sendRequest(prompt)
	if err != nil {
		return nil, err
	}

	// Parse response using the original texts count
	parsed, err := ParseResponse(result, len(texts))
	if err != nil {
		return nil, fmt.Errorf("error translating: %w", err)
	}
	return parsed, nil
}

func (o *OpenAI) Translate(texts []string, toLang string) ([]string, error) {
	return o.handler(texts, toLang)
}

func (o *OpenAI) OnTranslated(f func([]string) error) {
	o.onTrans = f
}

func (o *OpenAI) buildRequest(prompt string) (*http.Request, error) {
	// Build base request
	request := oai.ChatCompletionRequest{
		Model: o.model,
		Messages: []oai.ChatCompletionMessage{
			{
				Role:    oai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	// Convert request to map to merge req parameters
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	var reqMap map[string]any
	if err := json.Unmarshal(requestJSON, &reqMap); err != nil {
		return nil, fmt.Errorf("error unmarshaling request: %w", err)
	}

	// Merge all req configuration into request
	for key, value := range o.reqArgs {
		reqMap[key] = value
	}

	// Marshal final request
	jsonBody, err := json.Marshal(reqMap)
	if err != nil {
		return nil, fmt.Errorf("error marshaling final request: %w", err)
	}

	// Create HTTP request
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	return req, nil
}

func (o *OpenAI) sendRequest(prompt string) (string, error) {
	// Build request
	req, err := o.buildRequest(prompt)
	if err != nil {
		return "", err
	}

	// Make request using HTTP client
	resp, err := o.config.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response oai.ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	return response.Choices[0].Message.Content, nil
}

type Translation struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func ParseResponse(str string, expectedCount int) ([]string, error) {
	trans := []Translation{}
	if err := json.Unmarshal([]byte(str), &trans); err != nil {
		return nil, fmt.Errorf("error parsing response: %w, response str: %s", transerrors.ErrInvalidJSON, str)
	}

	// 验证返回的段数是否与期望的段数匹配
	if len(trans) != expectedCount {
		return nil, fmt.Errorf("error parsing response: %w, expected %d, got %d", transerrors.ErrCountMismatch, expectedCount, len(trans))
	}

	arr := make([]string, len(trans))
	for _, t := range trans {
		if t.ID >= len(arr) {
			return nil, fmt.Errorf("error parsing response: %w, invalid id %d", transerrors.ErrInvalidJSON, t.ID)
		}
		arr[t.ID] = t.Text
	}
	return arr, nil
}
