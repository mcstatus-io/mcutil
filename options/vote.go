package options

import "time"

// Vote is the options used by the SendVote() function
type Vote struct {
	ServiceName string
	Username    string
	Token       string
	UUID        string
	Timestamp   time.Time
	Timeout     time.Duration
}
