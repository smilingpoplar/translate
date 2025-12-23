package middleware

import (
	"fmt"
	"strings"
)

func TextsLimit(maxLen int) Middleware {
	return func(handler Handler) Handler {
		handler = TextsRegroup(maxLen)(handler)

		return func(texts []string, toLang string) ([]string, error) {
			texts, info, err := splitLongTexts(texts, maxLen)
			if err != nil {
				return nil, fmt.Errorf("error split long text: %w", err)
			}
			result, err := handler(texts, toLang)
			if err != nil {
				return nil, err
			}
			result = mergeBack(result, info)
			return result, nil
		}
	}
}

// 防止texts的总字节数>maxLen，对texts分组
func regroupTexts(texts []string, maxLen int) ([][]string, error) {
	var result [][]string
	var list []string
	currLen := 0
	for _, text := range texts {
		textLen := len(text)
		if textLen > maxLen {
			return nil, fmt.Errorf("text too long: textLen %d > maxLen %d, text: %s",
				textLen, maxLen, text)
		}

		if currLen+textLen <= maxLen {
			list = append(list, text)
			currLen += textLen
		} else {
			if currLen > 0 {
				result = append(result, list)
			}
			list = []string{text}
			currLen = textLen
		}
	}
	result = append(result, list)

	return result, nil
}

// 防止单text的字节数>maxLen，将该text拆分并追加到texts，同时将text清空
// 等翻译后，将text拆分翻译后的结果拼接，放回原位
type splitInfo struct {
	Len     int
	Mapping map[int][]int
}

func splitLongTexts(texts []string, maxLen int) ([]string, *splitInfo, error) {
	info := &splitInfo{Len: len(texts), Mapping: make(map[int][]int)}
	for i, text := range texts {
		if len(text) <= maxLen {
			continue
		}

		lines := splitLongText(text, maxLen)
		lists, err := regroupTexts(lines, maxLen)
		if err != nil {
			return nil, nil, err
		}
		for _, list := range lists {
			info.Mapping[i] = append(info.Mapping[i], len(texts))
			texts = append(texts, strings.Join(list, "\n"))
		}
		texts[i] = ""
	}
	return texts, info, nil
}

func splitLongText(text string, maxLen int) []string {
	var result []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if len(line) <= maxLen {
			result = append(result, line)
			continue
		}

		var paragram []string
		paragramLen := 0
		sentences := strings.Split(line, ". ")
		for i, sentence := range sentences {
			if i < len(sentences)-1 {
				sentence += ". "
			}
			sentenceLen := len(sentence)
			if paragramLen+sentenceLen <= maxLen {
				paragram = append(paragram, sentence)
				paragramLen += sentenceLen
			} else {
				if paragramLen > 0 {
					result = append(result, strings.Join(paragram, ""))
				}
				paragram = []string{sentence}
				paragramLen = sentenceLen
			}
		}
		result = append(result, strings.Join(paragram, ""))
	}
	return result
}

func mergeBack(list []string, info *splitInfo) []string {
	for i := range info.Mapping {
		result := make([]string, 0, len(info.Mapping[i]))
		for j := range info.Mapping[i] {
			result = append(result, list[info.Mapping[i][j]])
		}
		list[i] = strings.Join(result, "\n")
	}
	return list[:info.Len]
}
