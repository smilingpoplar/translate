package transerrors

import "errors"

var ErrInvalidJSON = errors.New("invalid json response")
var ErrTooManyRequests = errors.New("too many requests")
var ErrCountMismatch = errors.New("translation count mismatch")
var ErrNoTranslation = errors.New("no translation")
