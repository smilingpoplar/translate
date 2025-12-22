package middleware

import (
	"fmt"
	"sync"
)

// Concurrent 并发处理中间件，将文本分组并发处理
func Concurrent(maxConcurrency int) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			if len(texts) <= 1 || maxConcurrency <= 1 {
				// 如果文本数量少或并发数为1，直接顺序处理
				return handler(texts, toLang)
			}

			// 计算实际需要的组数
			groupCount := maxConcurrency
			if groupCount > len(texts) {
				groupCount = len(texts)
			}
			groupSize := (len(texts) + groupCount - 1) / groupCount // 向上取整，确保所有文本都被覆盖

			var wg sync.WaitGroup
			results := make([][]string, groupCount)
			errors := make([]error, groupCount)

			start := 0
			for i := 0; i < groupCount; i++ {
				end := start + groupSize
				if end > len(texts) {
					end = len(texts)
				}
				if start >= len(texts) {
					break // 没有更多文本需要处理
				}

				group := texts[start:end]
				wg.Add(1)

				go func(idx int, g []string) {
					defer wg.Done()
					result, err := handler(g, toLang)
					results[idx] = result
					errors[idx] = err
				}(i, group)

				start = end
			}

			wg.Wait()

			// 检查是否有错误
			for _, err := range errors {
				if err != nil {
					return nil, err
				}
			}

			// 合并结果
			var finalResult []string
			for i := 0; i < groupCount && i < len(results); i++ {
				finalResult = append(finalResult, results[i]...)
			}

			// 检查结果数量是否与原始数量匹配
			if len(finalResult) != len(texts) {
				return nil, fmt.Errorf("concurrent processing length mismatch: expected %d, got %d", len(texts), len(finalResult))
			}

			return finalResult, nil
		}
	}
}