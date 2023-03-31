package options

import (
	"time"

	"github.com/mcstatus-io/mcutil/description"
)

// JavaStatus is the options used by the Status() function
type JavaStatus struct {
	EnableSRV        bool
	Timeout          time.Duration
	ProtocolVersion  int
	DefaultMOTDColor description.Color
}

// JavaStatusLegacy is the options used by the StatusLegacy() function
type JavaStatusLegacy struct {
	EnableSRV        bool
	Timeout          time.Duration
	ProtocolVersion  int
	DefaultMOTDColor description.Color
}

// BedrockStatus is the options used by the StatusBedrock() function
type BedrockStatus struct {
	EnableSRV        bool
	Timeout          time.Duration
	ClientGUID       int64
	DefaultMOTDColor description.Color
}
