package storage

import "errors"

// Storage errors
var (
	ErrEmptyDelta  = errors.New("delta is empty")
	ErrNotFound    = errors.New("value not found")
	ErrUnknownType = errors.New("unknown type")
	ErrEmptyValue  = errors.New("value is empty")
)
