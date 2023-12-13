package openai

import (
	"context"
	"fmt"
	"strings"

	oai "github.com/sashabaranov/go-openai"
	"github.com/smilingpoplar/translate/translator"
	"github.com/smilingpoplar/translate/translator/decorator"
	"github.com/smilingpoplar/translate/util"
)

const BaseURL = "https://api.openai.com/v1"

type OpenAI struct {
	client *oai.Client
	inner  translator.Translator
}

func New(key, baseURL string) *OpenAI {
	o, _ := NewWithProxy(key, baseURL, "")
	return o
}

func NewWithProxy(key, baseURL, proxy string) (*OpenAI, error) {
	if baseURL == "" {
		baseURL = BaseURL
	} else {
		if strings.LastIndex(baseURL, "/v") == -1 {
			idx := strings.LastIndex(BaseURL, "/v")
			baseURL += BaseURL[idx:]
		}
	}

	o := &OpenAI{}
	config := oai.DefaultConfig(key)
	config.BaseURL = baseURL
	if proxy != "" {
		if err := util.SetProxy(proxy, config.HTTPClient); err != nil {
			return nil, fmt.Errorf("error creating openai translator: %w", err)
		}
	}
	o.client = oai.NewClientWithConfig(config)

	var fn translator.Translator = translator.TextTranslator(o.translate)
	fn = decorator.RetryDecorator(fn, 5, 3)
	o.inner = decorator.TextLimitDecorator(fn, 3000)
	return o, nil
}

func (o *OpenAI) Translate(texts []string, toLang string) ([]string, error) {
	return o.inner.Translate(texts, toLang)
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

func (o *OpenAI) OnTranslated(f func([]string) error) {
	o.inner.(*decorator.TextLimit).OnTranslated(f)
}
