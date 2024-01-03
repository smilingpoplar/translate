package openai

import (
	"context"
	"fmt"
	"strings"

	oai "github.com/sashabaranov/go-openai"
	"github.com/smilingpoplar/translate/translator/middleware"
	"github.com/smilingpoplar/translate/util"
)

const BaseURL = "https://api.openai.com/v1"

type OpenAI struct {
	config  *oai.ClientConfig
	client  *oai.Client
	handler middleware.Handler
	onTrans func([]string) error
}

type option func(*OpenAI) error

func New(key, baseURL string, opts ...option) (*OpenAI, error) {
	if baseURL == "" {
		baseURL = BaseURL
	} else {
		if strings.LastIndex(baseURL, "/v") == -1 {
			idx := strings.LastIndex(BaseURL, "/v")
			baseURL += BaseURL[idx:]
		}
	}

	o := &OpenAI{}
	chain := middleware.Chain(
		middleware.TextLimit(3000),
		middleware.OnTranslated(&o.onTrans),
		middleware.Retry(5, 3),
	)
	o.handler = chain(middleware.TextHandler(o.translate))

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

func (o *OpenAI) translate(text string, toLang string) (string, error) {
	prompt := fmt.Sprintf("You're a translator. Translate to %s.", toLang)
	resp, err := o.client.CreateChatCompletion(context.Background(), oai.ChatCompletionRequest{
		Model: oai.GPT3Dot5Turbo,
		Messages: []oai.ChatCompletionMessage{
			{
				Role:    oai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    oai.ChatMessageRoleUser,
				Content: text,
			},
		},
	},
	)
	if err != nil {
		return "", fmt.Errorf("error translating: %w", err)
	}

	result := resp.Choices[0].Message.Content
	return result, nil
}

func (o *OpenAI) Translate(texts []string, toLang string) ([]string, error) {
	return o.handler(texts, toLang)
}

func (o *OpenAI) OnTranslated(f func([]string) error) {
	o.onTrans = f
}
