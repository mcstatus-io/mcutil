package response

import (
	"time"

	"github.com/mcstatus-io/mcutil/v4/formatting"
)

type JavaStatus struct {
	Version   Version           `json:"version"`
	Players   Players           `json:"players"`
	MOTD      formatting.Result `json:"motd"`
	Favicon   *string           `json:"favicon"`
	SRVRecord *SRVRecord        `json:"srv_record"`
	Mods      *ModInfo          `json:"mods"`
	Latency   time.Duration     `json:"-"`
}

type Players struct {
	Max    *int64         `json:"max"`
	Online *int64         `json:"online"`
	Sample []SamplePlayer `json:"sample"`
}

type SamplePlayer struct {
	ID   string            `json:"id"`
	Name formatting.Result `json:"name"`
}

type ModInfo struct {
	Type string `json:"type"`
	List []Mod  `json:"list"`
}

type Mod struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type Version struct {
	Name     formatting.Result `json:"name"`
	Protocol int64             `json:"protocol"`
}

type JavaStatusLegacy struct {
	Version   *Version          `json:"version"`
	Players   LegacyPlayers     `json:"players"`
	MOTD      formatting.Result `json:"motd"`
	SRVRecord *SRVRecord        `json:"srv_record"`
}

type LegacyPlayers struct {
	Online int64 `json:"online"`
	Max    int64 `json:"max"`
}

type BedrockStatus struct {
	ServerGUID      int64              `json:"server_guid"`
	Edition         *string            `json:"edition"`
	MOTD            *formatting.Result `json:"motd"`
	ProtocolVersion *int64             `json:"protocol_version"`
	Version         *string            `json:"version"`
	OnlinePlayers   *int64             `json:"online_players"`
	MaxPlayers      *int64             `json:"max_players"`
	ServerID        *string            `json:"server_id"`
	Gamemode        *string            `json:"gamemode"`
	GamemodeID      *int64             `json:"gamemode_id"`
	PortIPv4        *uint16            `json:"port_ipv4"`
	PortIPv6        *uint16            `json:"port_ipv6"`
}
