package response

import "github.com/mcstatus-io/mcutil/v4/formatting"

// BasicQuery is the response data returned from doing a basic query on a server.
type BasicQuery struct {
	MOTD          formatting.Result `json:"motd"`
	GameType      string            `json:"game_type"`
	Map           string            `json:"map"`
	OnlinePlayers uint64            `json:"online_players"`
	MaxPlayers    uint64            `json:"max_players"`
	HostPort      uint16            `json:"host_port"`
	HostIP        string            `json:"host_ip"`
}

// FullQuery is the response data returned from doing a full query on a server.
type FullQuery struct {
	Data    map[string]string `json:"data"`
	Players []string          `json:"players"`
}
