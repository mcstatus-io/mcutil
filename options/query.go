package options

import "time"

// Query is the options used all query functions
type Query struct {
	Timeout   time.Duration
	SessionID int32
}
