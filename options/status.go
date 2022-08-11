package options

import "time"

type JavaStatus struct {
	EnableSRV       bool
	Timeout         time.Duration
	ProtocolVersion int
}

type JavaStatusLegacy struct {
	EnableSRV       bool
	Timeout         time.Duration
	ProtocolVersion int
}

type BedrockStatus struct {
	EnableSRV  bool
	Timeout    time.Duration
	ClientGUID int64
}
