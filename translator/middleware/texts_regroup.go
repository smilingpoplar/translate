package middleware

import (
	"fmt"
)

func TextsRegroup(maxLen int) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			groups, err := regroupTexts(texts, maxLen)
			if err != nil {
				return nil, fmt.Errorf("error group texts: %w", err)
			}

			var result []string
			for _, group := range groups {
				part, err := handler(group, toLang)
				if err != nil {
					return nil, err
				}
				result = append(result, part...)
			}
			return result, nil
		}
	}
}
