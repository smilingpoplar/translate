package middleware

import (
	"fmt"
	"strings"
)

// 将所有texts合并成一段，发往只接收text的翻译接口
func TextLimit(maxLen int) Middleware {
	return func(handler Handler) Handler {
		handler = TextsRegroup(maxLen)(handler)
		return func(texts []string, toLang string) ([]string, error) {
			texts, info, err := splitLongTexts([]string{strings.Join(texts, "\n")}, maxLen-len(texts)+1)
			if err != nil || len(info.Mapping) > 1 {
				return nil, fmt.Errorf("error split long text: %w", err)
			}
			if len(info.Mapping) == 1 {
				texts = texts[1:]
			}
			return handler(texts, toLang)
		}
	}
}
