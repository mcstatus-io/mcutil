package options

import (
	"time"

	"github.com/mcstatus-io/mcutil/description"
)

type JavaStatus struct {
	EnableSRV        bool
	Timeout          time.Duration
	ProtocolVersion  int
	DefaultMOTDColor description.Color
}

type JavaStatusLegacy struct {
	EnableSRV        bool
	Timeout          time.Duration
	ProtocolVersion  int
	DefaultMOTDColor description.Color
}

type BedrockStatus struct {
	EnableSRV        bool
	Timeout          time.Duration
	ClientGUID       int64
	DefaultMOTDColor description.Color
}
