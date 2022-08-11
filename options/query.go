package options

import "time"

type Query struct {
	Timeout   time.Duration
	SessionID int32
}
