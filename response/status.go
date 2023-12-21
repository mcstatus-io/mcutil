package response

import (
	"time"

	"github.com/mcstatus-io/mcutil/v3/formatting"
)

type JavaStatus struct {
	Version   Version           `json:"version"`
	Players   Players           `json:"players"`
	MOTD      formatting.Result `json:"motd"`
	Favicon   *string           `json:"favicon"`
	SRVResult *SRVRecord        `json:"srv_result"`
	ModInfo   *ModInfo          `json:"mod_info"`
	Latency   time.Duration     `json:"-"`
}

type Players struct {
	Max    *int64         `json:"max"`
	Online *int64         `json:"online"`
	Sample []SamplePlayer `json:"sample"`
}

type SamplePlayer struct {
	ID        string `json:"id"`
	NameRaw   string `json:"name_raw"`
	NameClean string `json:"name_clean"`
	NameHTML  string `json:"name_html"`
}

type ModInfo struct {
	Type string `json:"type"`
	Mods []Mod  `json:"mods"`
}

type Mod struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type Version struct {
	NameRaw   string `json:"name_raw"`
	NameClean string `json:"name_clean"`
	NameHTML  string `json:"name_html"`
	Protocol  int64  `json:"protocol"`
}

type JavaStatusLegacy struct {
	Version   *Version          `json:"version"`
	Players   LegacyPlayers     `json:"players"`
	MOTD      formatting.Result `json:"motd"`
	SRVResult *SRVRecord        `json:"srv_result"`
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
	SRVResult       *SRVRecord         `json:"srv_result"`
}
