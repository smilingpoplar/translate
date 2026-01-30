package middleware

import (
	"regexp"
	"sort"
	"strings"

	"github.com/smilingpoplar/translate/util"
)

func Glossary(terms map[string]string) Middleware {
	if len(terms) == 0 {
		return func(next Handler) Handler {
			return next
		}
	}

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
			placeholderMap := make(map[string]string)           // 占位符 => 译文
			translationToPlaceholder := make(map[string]string) // 译文 => 占位符
			nextID := 0
			textsWithPlaceholders := make([]string, len(texts))

			for i, text := range texts {
				processedText := text

				for _, term := range termList {
					if term.regex.MatchString(processedText) {
						placeholder, exists := translationToPlaceholder[term.to]
						if !exists {
							placeholder = util.GeneratePlaceholder(nextID)
							translationToPlaceholder[term.to] = placeholder
							placeholderMap[placeholder] = term.to
							nextID++
						}
						processedText = term.regex.ReplaceAllString(processedText, placeholder)
					}
				}

				textsWithPlaceholders[i] = processedText
			}

			// 阶段2：翻译（调用下一个中间件）
			result, err := handler(textsWithPlaceholders, toLang)
			if err != nil {
				return nil, err
			}

			// 阶段3：替换占位符为译文
			for i := range result {
				translatedText := result[i]
				for placeholder, target := range placeholderMap {
					translatedText = strings.ReplaceAll(translatedText, placeholder, target)
				}
				result[i] = translatedText
			}

			return result, nil
		}
	}
}
