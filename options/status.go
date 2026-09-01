package options

import (
	"time"
)

// StatusModern is the options used by the status.Modern() function.
type StatusModern struct {
	EnableSRV         bool
	Timeout           time.Duration
	ProtocolVersion   int
	ReceiveLimitBytes int64
	Ping              bool
	Debug             bool
}

// StatusLegacy is the options used by the status.Legacy() function.
type StatusLegacy struct {
	EnableSRV         bool
	Timeout           time.Duration
	ProtocolVersion   int
	ReceiveLimitBytes int64
}

// StatusBedrock is the options used by the status.Bedrock() function.
type StatusBedrock struct {
	Timeout           time.Duration
	ClientGUID        int64
	ReceiveLimitBytes int64
}
