package translator

import (
	"github.com/smilingpoplar/translate/config"
	"github.com/smilingpoplar/translate/translator/google"
	"github.com/smilingpoplar/translate/translator/openai"
)

const (
	kGoogle = "google"
	kOpenAI = "openai"
)

func GetTranslator(service, proxy string) (Translator, error) {
	sc := config.NewServiceConfig(service)
	var trans Translator
	var err error
	if sc.Name == kGoogle {
		trans, err = google.New(google.WithProxy(proxy))
	} else if sc.Name == kOpenAI || sc.Type == kOpenAI {
		trans, err = getTranslatorOpenAI(sc, proxy)
	}
	return trans, err
}

func getTranslatorOpenAI(sc *config.ServiceConfig, proxy string) (Translator, error) {
	if err := sc.ValidateEnvArgs(); err != nil {
		return nil, err
	}

	return openai.New(
		sc.GetEnvValue("api-key"),
		openai.WithBaseURL(sc.GetEnvValue("base-url")),
		openai.WithModel(sc.GetEnvValue("model")),
		openai.WithPrompt(sc.GetEnvValue("prompt")),
		openai.WithProxy(proxy),
	)
}
