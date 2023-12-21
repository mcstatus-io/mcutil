package response

import "github.com/mcstatus-io/mcutil/v3/formatting"

type BasicQuery struct {
	MOTD          formatting.Result `json:"motd"`
	GameType      string            `json:"game_type"`
	Map           string            `json:"map"`
	OnlinePlayers uint64            `json:"online_players"`
	MaxPlayers    uint64            `json:"max_players"`
	HostPort      uint16            `json:"host_port"`
	HostIP        string            `json:"host_ip"`
}

type FullQuery struct {
	Data    map[string]string `json:"data"`
	Players []string          `json:"players"`
}
