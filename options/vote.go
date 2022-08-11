package options

import "time"

type Vote struct {
	ServiceName string
	Username    string
	Token       string
	UUID        string
	Timestamp   time.Time
	Timeout     time.Duration
}
