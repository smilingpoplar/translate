package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"

	oai "github.com/sashabaranov/go-openai"
	"github.com/smilingpoplar/translate/config"
	"github.com/smilingpoplar/translate/translator/middleware"
	"github.com/smilingpoplar/translate/translator/transerrors"
	"github.com/smilingpoplar/translate/util"
)

type OpenAI struct {
	config      *oai.ClientConfig
	client      *oai.Client
	model       string
	prompt      string
	handler     middleware.Handler
	glossary    map[string]string
	onTrans     func([]string) error
	Name        string
	apiKey      string
	requestArgs map[string]any
}

type option func(*OpenAI) error

func New(sc *config.ServiceConfig, opts ...option) (*OpenAI, error) {
	name := sc.Name
	model := sc.GetEnvValue("model")
	key := sc.GetEnvValue("api-key")
	baseURL := sc.GetEnvValue("base-url")

	o := &OpenAI{Name: name, model: model}
	config := oai.DefaultConfig(key)
	config.BaseURL = baseURL
	o.config = &config
	o.client = oai.NewClientWithConfig(config)

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, fmt.Errorf("error creating openai translator: %w", err)
		}
	}

	rpm := sc.GetRpm()
	maxConcurrency := sc.GetMaxConcurrency()

	chain := middleware.Chain(
		middleware.TextsLimit(2000),
		middleware.OnTranslated(&o.onTrans),
		middleware.Glossary(o.glossary),
		middleware.Retry(8, 3),
		middleware.Cache(name),
		middleware.RateLimit(rpm),
		middleware.Concurrent(maxConcurrency),
	)
	o.handler = chain(o.translate)

	// 构造request时使用
	o.apiKey = key
	o.requestArgs = sc.GetRequestArgs()

	return o, nil
}

func WithProxy(proxy string) option {
	return func(o *OpenAI) error {
		return util.SetProxy(proxy, o.config.HTTPClient.(*http.Client))
	}
}

func WithGlossary(glossary map[string]string) option {
	return func(o *OpenAI) error {
		o.glossary = glossary
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
	for i := range texts {
		if texts[i] == parsed[i] && len(texts[i]) > 20 {
			return nil, transerrors.ErrNoTranslation
		}
	}
	return parsed, nil
}

func (o *OpenAI) Translate(texts []string, toLang string) ([]string, error) {
	return o.handler(texts, toLang)
}

func (o *OpenAI) OnTranslated(f func([]string) error) {
	o.onTrans = f
}

func (o *OpenAI) sendRequest(prompt string) (string, error) {
	request := oai.ChatCompletionRequest{
		Model: o.model,
		Messages: []oai.ChatCompletionMessage{
			{
				Role:    oai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	// 如果没有额外参数，直接使用库方法
	if len(o.requestArgs) == 0 {
		ctx := context.Background()
		response, err := o.client.CreateChatCompletion(ctx, request)
		if err != nil {
			return "", fmt.Errorf("error making request: %w", err)
		}
		return response.Choices[0].Message.Content, nil
	}

	// 有额外参数，使用手动构造方式
	return o.sendRequestWithExtra(request)
}

func (o *OpenAI) sendRequestWithExtra(request oai.ChatCompletionRequest) (string, error) {
	// 序列化request
	reqJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// 合并额外参数requestArgs
	var reqMap map[string]any
	if err := json.Unmarshal(reqJSON, &reqMap); err != nil {
		return "", err
	}
	maps.Copy(reqMap, o.requestArgs)
	finalJSON, err := json.Marshal(reqMap)
	if err != nil {
		return "", err
	}

	// 构造并发送 HTTP 请求
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST",
		o.config.BaseURL+"/chat/completions", bytes.NewReader(finalJSON))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.config.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var response oai.ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
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
