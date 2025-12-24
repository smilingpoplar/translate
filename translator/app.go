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

func GetTranslator(service, proxy string, glossary map[string]string) (Translator, error) {
	sc := config.NewServiceConfig(service)
	var trans Translator
	var err error
	if sc.Name == kGoogle {
		trans, err = google.New(google.WithProxy(proxy), google.WithGlossary(glossary))
	} else if sc.Name == kOpenAI || sc.Type == kOpenAI {
		trans, err = getTranslatorOpenAI(sc, proxy, glossary)
	}
	return trans, err
}

func getTranslatorOpenAI(sc *config.ServiceConfig, proxy string, glossary map[string]string) (Translator, error) {
	if err := sc.ValidateEnvArgs(); err != nil {
		return nil, err
	}

	return openai.New(sc, openai.WithProxy(proxy), openai.WithGlossary(glossary))
}
