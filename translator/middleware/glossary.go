package middleware

import (
	"regexp"
	"sort"
	"strings"

	"github.com/smilingpoplar/translate/util"
)

var placeholderRegex = regexp.MustCompile(`(?i)\{id_\d+\}`)

func Glossary(terms map[string]string) Middleware {
	return func(handler Handler) Handler {
		type termInfo struct {
			from  string
			to    string
			regex *regexp.Regexp
		}

		termList := make([]termInfo, 0, len(terms))
		for from, to := range terms {
			regex, err := util.BuildWordBoundaryRegex(from)
			if err != nil {
				continue // 跳过无法编译的术语
			}
			termList = append(termList, termInfo{
				from:  from,
				to:    to,
				regex: regex,
			})
		}

		// 按长度降序排序（先匹配长术语）
		sort.Slice(termList, func(i, j int) bool {
			return len(termList[i].from) > len(termList[j].from)
		})

		return func(texts []string, toLang string) ([]string, error) {
			// 阶段1：替换原文为占位符
			termToPlaceholder := make(map[string]string)          // 术语原文 => 占位符
			placeholderToTranslation := make(map[string]string)   // 占位符 => 译文
			nextID := 0
			textsWithPlaceholders := make([]string, len(texts))
			sourceTranslationsByText := make([][]string, len(texts))

			for i, text := range texts {
				processedText := text

				for _, term := range termList {
					if term.regex.MatchString(processedText) {
						placeholder, exists := termToPlaceholder[term.from]
						if !exists {
							placeholder = util.GeneratePlaceholder(nextID)
							nextID++
							termToPlaceholder[term.from] = placeholder
							placeholderToTranslation[canonicalPlaceholder(placeholder)] = term.to
						}
						processedText = term.regex.ReplaceAllString(processedText, placeholder)
					}
				}

				textsWithPlaceholders[i] = processedText
				sourceTranslationsByText[i] = collectSourceTranslations(processedText, placeholderToTranslation)
			}

			// 阶段2：翻译（调用下一个中间件）
			result, err := handler(textsWithPlaceholders, toLang)
			if err != nil {
				return nil, err
			}

			// 阶段3：替换占位符为译文
			for i := range result {
				result[i] = replacePlaceholdersByOrder(result[i], sourceTranslationsByText[i])
			}

			return result, nil
		}
	}
}

func replacePlaceholdersByOrder(translatedText string, sourceTranslations []string) string {
	index := 0
	return placeholderRegex.ReplaceAllStringFunc(translatedText, func(token string) string {
		if index >= len(sourceTranslations) {
			return ""
		}
		translation := sourceTranslations[index]
		index++
		return translation
	})
}

func collectSourceTranslations(sourceText string, placeholderToTranslation map[string]string) []string {
	tokens := placeholderRegex.FindAllString(sourceText, -1)
	translations := make([]string, 0, len(tokens))
	for _, token := range tokens {
		translation, ok := placeholderToTranslation[canonicalPlaceholder(token)]
		if ok {
			translations = append(translations, translation)
			continue
		}
		// source 原本就存在的 {ID_n}（非 glossary 注入）按原样保留
		translations = append(translations, token)
	}
	return translations
}

func canonicalPlaceholder(token string) string {
	return strings.ToUpper(token)
}
