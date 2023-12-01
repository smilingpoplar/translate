package util

import (
	"errors"
	"time"
)

var ErrTooManyRequests = errors.New("too many requests")

func Retry(retryOp func() ([]string, error)) ([]string, error) {
	var result []string
	var err error

	baseDelay := 5 // seconds
	retryCount := 5
	for i := 1; i <= retryCount; i++ {
		result, err = retryOp()
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
