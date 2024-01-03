package middleware

import (
	"errors"
	"time"
)

var ErrTooManyRequests = errors.New("too many requests")

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
				if !errors.Is(err, ErrTooManyRequests) {
					return nil, err
				}

				time.Sleep(time.Duration(baseDelay * i))
			}

			return nil, err
		}
	}
}
