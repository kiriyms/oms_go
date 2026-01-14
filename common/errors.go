package common

import "errors"

var (
	ErrNoItems = errors.New("items must contain at least one item")
)
