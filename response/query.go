package response

import "github.com/mcstatus-io/mcutil/description"

type BasicQuery struct {
	MOTD          description.Formatting
	GameType      string
	Map           string
	OnlinePlayers uint64
	MaxPlayers    uint64
	HostPort      uint16
	HostIP        string
}

type FullQuery struct {
	Data    map[string]string
	Players []string
}
