package response

import "github.com/mcstatus-io/mcutil/v4/formatting"

// QueryBasic is the response data returned from doing a basic query on a server.
type QueryBasic struct {
	MOTD          formatting.Result `json:"motd"`
	GameType      string            `json:"game_type"`
	Map           string            `json:"map"`
	OnlinePlayers uint64            `json:"online_players"`
	MaxPlayers    uint64            `json:"max_players"`
	HostPort      uint16            `json:"host_port"`
	HostIP        string            `json:"host_ip"`
}

// QueryFull is the response data returned from doing a full query on a server.
type QueryFull struct {
	Data    map[string]string `json:"data"`
	Players []string          `json:"players"`
}
