package options

import "time"

// RCON is the options used when connecting using the RCON connection methods.
type RCON struct {
	Timeout time.Duration
}
