package options

import (
	"time"
)

// JavaStatus is the options used by the Status() function
type JavaStatus struct {
	EnableSRV       bool
	Timeout         time.Duration
	ProtocolVersion int
	Ping            bool
}

// JavaStatusLegacy is the options used by the StatusLegacy() function
type JavaStatusLegacy struct {
	EnableSRV       bool
	Timeout         time.Duration
	ProtocolVersion int
}

// BedrockStatus is the options used by the StatusBedrock() function
type BedrockStatus struct {
	EnableSRV  bool
	Timeout    time.Duration
	ClientGUID int64
}
