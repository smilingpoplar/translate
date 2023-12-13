package translator

import "strings"

type Translator interface {
	Translate(texts []string, toLang string) ([]string, error)
}

type TextsTranslator func(texts []string, toLang string) ([]string, error)

func (fn TextsTranslator) Translate(texts []string, toLang string) ([]string, error) {
	return fn(texts, toLang)
}

type TextTranslator func(text string, toLang string) (string, error)

func (fn TextTranslator) Translate(texts []string, toLang string) ([]string, error) {
	result, error := fn(strings.Join(texts, "\n"), toLang)
	return []string{result}, error
}

type TranslationObserver interface {
	OnTranslated(func([]string) error)
}
