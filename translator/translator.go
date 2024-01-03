package translator

type Translator interface {
	Translate(texts []string, toLang string) ([]string, error)
}

type TranslationObserver interface {
	OnTranslated(func([]string) error)
}
