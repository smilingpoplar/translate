package util

import "fmt"

// 防止texts的总字节数>maxLen，对texts分组
func RegroupTexts(texts []string, maxLen int) ([][]string, error) {
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
