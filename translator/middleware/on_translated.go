package middleware

func OnTranslated(onTrans *func([]string) error) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			result, err := handler(texts, toLang)
			if err != nil {
				return nil, err
			}
			if *onTrans != nil {
				if err = (*onTrans)(result); err != nil {
					return nil, err
				}
			}
			return result, nil
		}
	}
}
