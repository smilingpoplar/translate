package middleware

import "strings"

type Handler func(texts []string, toLang string) ([]string, error)
type Middleware func(Handler) Handler

func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}

func TextHandler(fn func(string, string) (string, error)) Handler {
	return func(texts []string, toLang string) ([]string, error) {
		result, err := fn(strings.Join(texts, "\n"), toLang)
		return []string{result}, err
	}
}
