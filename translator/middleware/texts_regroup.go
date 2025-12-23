package middleware

import (
	"fmt"
	"sync"
)

func TextsRegroup(maxLen int) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			groups, err := regroupTexts(texts, maxLen)
			if err != nil {
				return nil, fmt.Errorf("error group texts: %w", err)
			}

			results := make([][]string, len(groups))
			errs := make([]error, len(groups))
			var wg sync.WaitGroup
			for i, group := range groups {
				wg.Add(1)
				go func(index int, g []string) {
					defer wg.Done()
					results[index], errs[index] = handler(g, toLang)
				}(i, group)
			}
			wg.Wait()

			for _, err := range errs {
				if err != nil {
					return nil, err
				}
			}

			var finalResult []string
			for _, res := range results {
				finalResult = append(finalResult, res...)
			}
			return finalResult, nil
		}
	}
}
