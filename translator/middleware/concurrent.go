package middleware

func Concurrent(maxConcurrency int) Middleware {
	if maxConcurrency <= 0 {
		return func(handler Handler) Handler {
			return handler
		}
	}

	semaphore := make(chan struct{}, maxConcurrency)
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			return handler(texts, toLang)
		}
	}
}
