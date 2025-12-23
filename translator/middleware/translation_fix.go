package middleware

import (
	"strings"
)

func TranslationFix(fixes map[string]string) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			result, err := handler(texts, toLang)
			if err != nil {
				return nil, err
			}
			if fixes != nil {
				applyFixes(result, fixes)
			}
			return result, nil
		}
	}
}

func applyFixes(texts []string, fixes map[string]string) {
	for from, to := range fixes {
		for i := range texts {
			texts[i] = strings.ReplaceAll(texts[i], from, to)
		}
	}
}
