package options

import "time"

// Vote is the options used by the SendVote() function
type Vote struct {
	// Deprecated: This property no longer affects how the vote is sent or processed.
	RequireVersion int
	PublicKey      string
	ServiceName    string
	Username       string
	Token          string
	UUID           string
	IPAddress      string
	Timestamp      time.Time
	Timeout        time.Duration
}
