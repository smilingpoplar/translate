package middleware

import (
	"errors"
	"time"

	"github.com/smilingpoplar/translate/translator/transerrors"
	"github.com/smilingpoplar/translate/util"
)

// Retry 简单重试中间件（不使用缓存）
func Retry(baseDelay, retryCount int) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			var result []string
			var err error

			for i := 1; i <= retryCount; i++ {
				result, err = handler(texts, toLang)
				if err == nil {
					return result, nil
				}

				if !isRetryable(err) {
					return nil, err
				}

				time.Sleep(time.Duration(baseDelay*i) * time.Second)
			}

			return nil, err
		}
	}
}

// RetryWithCache 带缓存的重试中间件
func RetryWithCache(serviceName string, baseDelay, retryCount int) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			return retryWithCache(handler, serviceName, toLang, texts, baseDelay, retryCount)
		}
	}
}

// retryWithCache 带缓存的重试
func retryWithCache(handler Handler, serviceName, toLang string, texts []string, baseDelay, retryCount int) ([]string, error) {
	results := make([]string, len(texts))

	// 重试翻译，每次重试都从全局缓存获取最新翻译
	for retry := 1; retry <= retryCount; retry++ {
		// 每次重试都重新检查缓存，收集未缓存的文本
		uncached := make(map[int]string) // 索引 => 文本
		for i, text := range texts {
			if cached, found := util.Get(serviceName, toLang, text); found {
				results[i] = cached
			} else {
				uncached[i] = text
			}
		}

		// 所有文本都已缓存，直接返回
		if len(uncached) == 0 {
			return results, nil
		}

		// 构建待翻译文本列表
		indices := make([]int, 0, len(uncached))
		textsToTranslate := make([]string, 0, len(uncached))
		for idx, text := range uncached {
			indices = append(indices, idx)
			textsToTranslate = append(textsToTranslate, text)
		}

		// 调用翻译服务
		translatedTexts, err := handler(textsToTranslate, toLang)

		if err == nil {
			// 缓存并合并已翻译的结果
			for i, translated := range translatedTexts {
				idx, text := indices[i], textsToTranslate[i]
				results[idx] = translated
				util.Set(serviceName, toLang, text, translated)
			}
			util.SaveCache()

			return results, nil
		}

		if !isRetryable(err) {
			return nil, err
		}

		time.Sleep(time.Duration(baseDelay*retry) * time.Second)
	}

	return nil, errors.New("max retries exceeded")
}

// isRetryable 判断错误是否可重试
func isRetryable(err error) bool {
	// 限流错误
	if errors.Is(err, transerrors.ErrTooManyRequests) {
		return true
	}
	// json响应错误
	if errors.Is(err, transerrors.ErrInvalidJSON) {
		return true
	}
	// 翻译数量不匹配错误（部分翻译）
	if errors.Is(err, transerrors.ErrCountMismatch) {
		return true
	}
	return false
}
