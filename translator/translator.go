package translator

type Translator interface {
	Translate(texts []string, toLang string) ([]string, error)
}

type TextsTranslator func(texts []string, toLang string) ([]string, error)

func (fn TextsTranslator) Translate(texts []string, toLang string) ([]string, error) {
	return fn(texts, toLang)
}
