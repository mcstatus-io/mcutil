package options

import "time"

// RCON is the options used when connecting in the RCON#Connect() method
type RCON struct {
	Timeout time.Duration
}
