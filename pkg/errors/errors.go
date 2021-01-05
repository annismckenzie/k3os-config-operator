package errors

import (
	"errors"
)

// Is reports whether any error in err's chain matches target.
// See documentation on errors.Is for more information.
var Is = errors.Is

// ErrSkipUpdate tells the caller that the update should be skipped.
var ErrSkipUpdate = errors.New("no updates required, skip")
