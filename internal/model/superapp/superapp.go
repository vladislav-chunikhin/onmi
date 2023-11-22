package superapp

import "errors"

// ErrBlocked reports if super app service is blocked.
var ErrBlocked = errors.New("blocked")

// Batch is a batch of items.
type Batch []Item

// Item is some abstract item.
type Item struct{}
