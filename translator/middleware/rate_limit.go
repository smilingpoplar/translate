package middleware

import (
	"context"

	"golang.org/x/time/rate"
)

func RateLimit(rpm int) Middleware {
	limiter := rate.NewLimiter(rate.Limit(rpm)/60, 6)

	return func(next Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			if err := limiter.Wait(context.Background()); err != nil {
				return nil, err
			}
			return next(texts, toLang)
		}
	}
}
