package mcutil

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/description"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
)

var (
	defaultBedrockStatusOptions = options.BedrockStatus{
		EnableSRV:  true,
		Timeout:    time.Second * 5,
		ClientGUID: 0,
	}
	bedrockMagic = []byte{0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78}
)

// StatusBedrock retrieves the status of a Bedrock Minecraft server
func StatusBedrock(host string, port uint16, options ...options.BedrockStatus) (*response.BedrockStatus, error) {
	opts := parseBedrockStatusOptions(options...)

	var srvResult *response.SRVRecord = nil

	if opts.EnableSRV {
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

	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	// Unconnected ping packet
	// https://wiki.vg/Raknet_Protocol#Unconnected_Ping
	{
		buf := &bytes.Buffer{}

		// Packet ID - byte
		if err := buf.WriteByte(0x01); err != nil {
			return nil, err
		}

		// Time - int64
		if err := binary.Write(buf, binary.BigEndian, time.Now().UnixNano()/int64(time.Millisecond)); err != nil {
			return nil, err
		}

		// Magic - bytes
		if _, err := buf.Write(bedrockMagic); err != nil {
			return nil, err
		}

		// Client GUID - int64
		if err := binary.Write(buf, binary.BigEndian, opts.ClientGUID); err != nil {
			return nil, err
		}

		if _, err := io.Copy(conn, buf); err != nil {
			return nil, err
		}
	}

	var serverGUID int64
	var serverID string

	// Unconnected pong packet
	// https://wiki.vg/Raknet_Protocol#Unconnected_Pong
	{
		// Type - byte
		{
			v, err := r.ReadByte()

			if err != nil {
				return nil, err
			}

			if v != 0x1C {
				return nil, ErrUnexpectedResponse
			}
		}

		// Time - int64
		{
			data := make([]byte, 8)

			if _, err := r.Read(data); err != nil {
				return nil, err
			}
		}

		// Server GUID - int64
		{
			if err := binary.Read(r, binary.BigEndian, &serverGUID); err != nil {
				return nil, err
			}
		}

		// Magic - bytes
		{
			data := make([]byte, 16)

			if _, err := r.Read(data); err != nil {
				return nil, err
			}
		}

		// Server ID - string
		{
			var length uint16

			if err := binary.Read(r, binary.BigEndian, &length); err != nil {
				return nil, err
			}

			data := make([]byte, length)

			if _, err = r.Read(data); err != nil {
				return nil, err
			}

			serverID = string(data)
		}
	}

	response := &response.BedrockStatus{
		ServerGUID:      serverGUID,
		Edition:         nil,
		MOTD:            nil,
		ProtocolVersion: nil,
		Version:         nil,
		OnlinePlayers:   nil,
		MaxPlayers:      nil,
		ServerID:        nil,
		Gamemode:        nil,
		GamemodeID:      nil,
		PortIPv4:        nil,
		PortIPv6:        nil,
		SRVResult:       srvResult,
	}

	splitID := strings.Split(serverID, ";")

	var motd string

	for k, v := range splitID {
		if len(strings.Trim(v, " ")) < 1 {
			continue
		}

		switch k {
		case 0:
			{
				response.Edition = &splitID[k]

				break
			}
		case 1:
			{
				motd = splitID[k]

				break
			}
		case 2:
			{
				protocolVersion, err := strconv.ParseInt(splitID[k], 10, 64)

				if err != nil {
					return nil, err
				}

				response.ProtocolVersion = &protocolVersion

				break
			}
		case 3:
			{
				response.Version = &splitID[k]

				break
			}
		case 4:
			{
				onlinePlayers, err := strconv.ParseInt(splitID[k], 10, 64)

				if err != nil {
					return nil, err
				}

				response.OnlinePlayers = &onlinePlayers

				break
			}
		case 5:
			{
				maxPlayers, err := strconv.ParseInt(splitID[k], 10, 64)

				if err != nil {
					return nil, err
				}

				response.MaxPlayers = &maxPlayers

				break
			}
		case 6:
			{
				response.ServerID = &splitID[k]

				break
			}
		case 7:
			{
				motd += "\n" + splitID[k]

				break
			}
		case 8:
			{
				response.Gamemode = &splitID[k]

				break
			}
		case 9:
			{
				gamemodeID, err := strconv.ParseInt(splitID[k], 10, 64)

				if err != nil {
					return nil, err
				}

				response.GamemodeID = &gamemodeID

				break
			}
		case 10:
			{
				portIPv4, err := strconv.ParseInt(splitID[k], 10, 64)

				if err != nil {
					return nil, err
				}

				convertedIPv4 := uint16(portIPv4)

				response.PortIPv4 = &convertedIPv4

				break
			}
		case 11:
			{
				portIPv6, err := strconv.ParseInt(splitID[k], 10, 64)

				if err != nil {
					return nil, err
				}

				convertedIPv6 := uint16(portIPv6)

				response.PortIPv6 = &convertedIPv6

				break
			}
		}
	}

	if len(motd) > 0 {
		parsedMOTD, err := description.ParseMOTD(motd)

		if err != nil {
			return nil, err
		}

		response.MOTD = parsedMOTD
	}

	return response, nil
}

func parseBedrockStatusOptions(opts ...options.BedrockStatus) options.BedrockStatus {
	if len(opts) < 1 {
		options := options.BedrockStatus(defaultBedrockStatusOptions)

		options.ClientGUID = 2

		return options
	}

	return opts[0]
}
