package response

import (
	"time"

	"github.com/mcstatus-io/mcutil/v4/formatting"
)

// StatusModern is the response data returned from performing a status lookup on a modern Minecraft Java Edition server.
type StatusModern struct {
	Version   Version           `json:"version"`
	Players   Players           `json:"players"`
	MOTD      formatting.Result `json:"motd"`
	Favicon   *string           `json:"favicon"`
	SRVRecord *SRVRecord        `json:"srv_record"`
	Mods      *ModInfo          `json:"mods"`
	Latency   time.Duration     `json:"-"`
}

// Players contains data about the players on the server.
type Players struct {
	Max    *int64         `json:"max"`
	Online *int64         `json:"online"`
	Sample []SamplePlayer `json:"sample"`
}

// SamplePlayer is a single player returned from sample player data of a server.
type SamplePlayer struct {
	ID   string            `json:"id"`
	Name formatting.Result `json:"name"`
}

// ModInfo is the mods information of a server.
type ModInfo struct {
	Type string `json:"type"`
	List []Mod  `json:"list"`
}

// Mod is a single mod returned in the mod information of a server.
type Mod struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// Version is versioning information about a Minecraft server.
type Version struct {
	Name     formatting.Result `json:"name"`
	Protocol int64             `json:"protocol"`
}

// StatusLegacy is the response data returned from performing a status lookup on a legacy Minecraft Java Edition server.
type StatusLegacy struct {
	Version   *Version          `json:"version"`
	Players   LegacyPlayers     `json:"players"`
	MOTD      formatting.Result `json:"motd"`
	SRVRecord *SRVRecord        `json:"srv_record"`
}

// LegacyPlayers is the player information returned from a legacy server. This is the
// same as `response.Players` but without sample player data.
type LegacyPlayers struct {
	Online int64 `json:"online"`
	Max    int64 `json:"max"`
}

// StatusBedrock is the response data returned from a Minecraft Bedrock Edition server.
type StatusBedrock struct {
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
