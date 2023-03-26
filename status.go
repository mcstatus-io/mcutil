package mcutil

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/description"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
)

var (
	defaultJavaStatusOptions = options.JavaStatus{
		EnableSRV:        true,
		Timeout:          time.Second * 5,
		ProtocolVersion:  47,
		DefaultMOTDColor: description.White,
	}
)

type rawJavaStatus struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    *int `json:"max"`
		Online *int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description interface{} `json:"description"`
	Favicon     *string     `json:"favicon"`
	ModInfo     struct {
		List []struct {
			ID      string `json:"modid"`
			Version string `json:"version"`
		} `json:"modList"`
		Type string `json:"type"`
	} `json:"modinfo"`
}

// Status retrieves the status of any Minecraft server
func Status(host string, port uint16, options ...options.JavaStatus) (*response.JavaStatus, error) {
	opts := parseJavaStatusOptions(options...)

	var srvResult *response.SRVRecord = nil

	if opts.EnableSRV && port == 25565 {
		record, err := LookupSRV(host, port)

		if err == nil && record != nil {
			host = record.Target
			port = record.Port

			srvResult = &response.SRVRecord{
				Host: record.Target,
				Port: record.Port,
			}
		}
	}

	conn, err := net.DialTimeout("tcp4", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	// Handshake packet
	// https://wiki.vg/Server_List_Ping#Handshake
	{
		buf := &bytes.Buffer{}

		// Packet ID - varint
		if _, err := writeVarInt(0x00, buf); err != nil {
			return nil, err
		}

		// Protocol version - varint
		if _, err = writeVarInt(int32(opts.ProtocolVersion), buf); err != nil {
			return nil, err
		}

		// Host - string
		if err := writeString(host, buf); err != nil {
			return nil, err
		}

		// Port - uint16
		if err := binary.Write(buf, binary.BigEndian, port); err != nil {
			return nil, err
		}

		// Next state - varint
		if _, err := writeVarInt(1, buf); err != nil {
			return nil, err
		}

		if err := writePacket(buf, conn); err != nil {
			return nil, err
		}
	}

	// Request packet
	// https://wiki.vg/Server_List_Ping#Request
	{
		buf := &bytes.Buffer{}

		// Packet ID - varint
		if _, err := writeVarInt(0x00, buf); err != nil {
			return nil, err
		}

		if err := writePacket(buf, conn); err != nil {
			return nil, err
		}
	}

	var rawResponse rawJavaStatus

	// Response packet
	// https://wiki.vg/Server_List_Ping#Response
	{
		// Packet length - varint
		{
			if _, _, err := readVarInt(r); err != nil {
				return nil, err
			}
		}

		// Packet type - varint
		{
			packetType, _, err := readVarInt(r)

			if err != nil {
				return nil, err
			}

			if packetType != 0x00 {
				return nil, ErrUnexpectedResponse
			}
		}

		// Data - string
		{
			data, err := readString(r)

			if err != nil {
				return nil, err
			}

			if err = json.Unmarshal(data, &rawResponse); err != nil {
				return nil, err
			}
		}
	}

	payload := rand.Int63()

	// Ping packet
	// https://wiki.vg/Server_List_Ping#Ping
	{
		buf := &bytes.Buffer{}

		// Packet ID - varint
		if _, err := writeVarInt(0x01, buf); err != nil {
			return nil, err
		}

		// Payload - int64
		if err := binary.Write(buf, binary.BigEndian, payload); err != nil {
			return nil, err
		}

		if err := writePacket(buf, conn); err != nil {
			return nil, err
		}
	}

	pingStart := time.Now()

	// Pong packet
	// https://wiki.vg/Server_List_Ping#Pong
	{
		// Packet length - varint
		{
			if _, _, err := readVarInt(r); err != nil {
				return nil, err
			}
		}

		// Packet type - varint
		{
			packetType, _, err := readVarInt(r)

			if err != nil {
				return nil, err
			}

			if packetType != 0x01 {
				return nil, ErrUnexpectedResponse
			}
		}

		// Payload - int64
		{
			var returnPayload int64

			if err := binary.Read(r, binary.BigEndian, &returnPayload); err != nil {
				return nil, err
			}

			if payload != returnPayload {
				return nil, ErrUnexpectedResponse
			}
		}
	}

	motd, err := description.ParseMOTD(rawResponse.Description, opts.DefaultMOTDColor)

	if err != nil {
		return nil, err
	}

	samplePlayers := make([]response.SamplePlayer, 0)

	if rawResponse.Players.Sample != nil {
		for _, player := range rawResponse.Players.Sample {
			name, err := description.ParseMOTD(player.Name)

			if err != nil {
				return nil, err
			}

			samplePlayers = append(samplePlayers, response.SamplePlayer{
				ID:        player.ID,
				NameRaw:   name.Raw,
				NameClean: name.Clean,
				NameHTML:  name.HTML,
			})
		}
	}

	version, err := description.ParseMOTD(rawResponse.Version.Name)

	if err != nil {
		return nil, err
	}

	result := &response.JavaStatus{
		Version: response.Version{
			NameRaw:   version.Raw,
			NameClean: version.Clean,
			NameHTML:  version.HTML,
			Protocol:  rawResponse.Version.Protocol,
		},
		Players: response.Players{
			Online: rawResponse.Players.Online,
			Max:    rawResponse.Players.Max,
			Sample: samplePlayers,
		},
		MOTD:      *motd,
		Favicon:   rawResponse.Favicon,
		SRVResult: srvResult,
		Latency:   time.Since(pingStart),
		ModInfo:   nil,
	}

	if len(rawResponse.ModInfo.Type) > 0 {
		mods := make([]response.Mod, 0)

		for _, mod := range rawResponse.ModInfo.List {
			mods = append(mods, response.Mod{
				ID:      mod.ID,
				Version: mod.Version,
			})
		}

		result.ModInfo = &response.ModInfo{
			Type: rawResponse.ModInfo.Type,
			Mods: mods,
		}
	}

	return result, nil
}

func parseJavaStatusOptions(opts ...options.JavaStatus) options.JavaStatus {
	if len(opts) < 1 {
		return defaultJavaStatusOptions
	}

	return opts[0]
}
