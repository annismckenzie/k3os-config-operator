package errors

import (
	"errors"
)

// Is reports whether any error in err's chain matches target.
// See documentation on errors.Is for more information.
var Is = errors.Is

// New returns an error that formats as the given text.
// See documentation on errors.New for more information.
var New = errors.New

// ErrSkipUpdate tells the caller that the update should be skipped.
var ErrSkipUpdate = errors.New("no updates required, skip")

// ErrNilObjectPassed is returned if the caller passed a nil object.
var ErrNilObjectPassed = errors.New("nil object was passed")
