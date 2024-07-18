package options

import "time"

// Query is the options used by all query functions.
type Query struct {
	Timeout   time.Duration
	SessionID int32
}
