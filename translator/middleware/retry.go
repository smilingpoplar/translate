package middleware

import (
	"errors"
	"time"

	"github.com/smilingpoplar/translate/translator/transerrors"
)

func Retry(retryCount, baseDelay int) Middleware {
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

			return nil, errors.New("max retries exceeded")
		}
	}
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
	// 没有翻译（原样返回）
	if errors.Is(err, transerrors.ErrNoTranslation) {
		return true
	}
	return false
}
