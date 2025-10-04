package openai

import (
	"context"
	"encoding/json"
	"fmt"

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
}

type option func(*OpenAI) error

func New(name, key, baseURL, model string, opts ...option) (*OpenAI, error) {
	o := &OpenAI{Name: name, model: model}
	chain := middleware.Chain(
		middleware.TextsLimit(2000),
		middleware.OnTranslated(&o.onTrans),
		middleware.RetryWithCache(name, 3, 8),
	)
	o.handler = chain(o.translate)

	config := oai.DefaultConfig(key)
	config.BaseURL = baseURL
	o.config = &config
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
		return util.SetProxy(proxy, o.config.HTTPClient)
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

	resp, err := o.client.CreateChatCompletion(context.Background(), oai.ChatCompletionRequest{
		Model: o.model,
		Messages: []oai.ChatCompletionMessage{
			{
				Role:    oai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	},
	)
	if err != nil {
		return nil, fmt.Errorf("error translating: %w", err)
	}

	result := resp.Choices[0].Message.Content
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
