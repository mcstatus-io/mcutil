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

// LegacyVote is the options used by the SendLegacyVote() function
type LegacyVote struct {
	PublicKey   string
	ServiceName string
	Username    string
	IPAddress   string
	Timestamp   time.Time
	Timeout     time.Duration
}
