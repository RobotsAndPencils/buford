// Package payload serializes a JSON payload to push.
package payload

import "errors"

// validation errors
var (
	ErrIncomplete = errors.New("payload does not contain necessary fields")
)
