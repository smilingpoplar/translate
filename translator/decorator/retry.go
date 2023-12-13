package decorator

import (
	"errors"
	"time"

	"github.com/smilingpoplar/translate/translator"
)

var ErrTooManyRequests = errors.New("too many requests")

func RetryDecorator(fn translator.Translator, baseDelay, retryCount int) translator.TextsTranslator {
	return func(texts []string, toLang string) ([]string, error) {
		var result []string
		var err error

		for i := 1; i <= retryCount; i++ {
			result, err = fn.Translate(texts, toLang)
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
