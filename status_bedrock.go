package mcutil

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/v3/formatting"
	"github.com/mcstatus-io/mcutil/v3/options"
	"github.com/mcstatus-io/mcutil/v3/response"
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
func StatusBedrock(ctx context.Context, host string, port uint16, options ...options.BedrockStatus) (*response.BedrockStatus, error) {
	r := make(chan *response.BedrockStatus, 1)
	e := make(chan error, 1)

	go func() {
		result, err := getStatusBedrock(host, port, options...)

		if err != nil {
			e <- err
		} else if result != nil {
			r <- result
		}
	}()

	select {
	case <-ctx.Done():
		if v := ctx.Err(); v != nil {
			return nil, v
		}

		return nil, context.DeadlineExceeded
	case v := <-r:
		return v, nil
	case v := <-e:
		return nil, v
	}
}

func getStatusBedrock(host string, port uint16, options ...options.BedrockStatus) (*response.BedrockStatus, error) {
	opts := parseBedrockStatusOptions(options...)

	var srvResult *response.SRVRecord = nil

	if opts.EnableSRV && port == 19132 {
		record, err := LookupSRV("udp", host)

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
		if err := binary.Write(buf, binary.BigEndian, time.Now().UnixMilli()); err != nil {
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
			var packetType byte

			if err := binary.Read(r, binary.BigEndian, &packetType); err != nil {
				return nil, err
			}

			if packetType != 0x1C {
				return nil, fmt.Errorf("statusbedrock: received unexpected packet type (expected=0x1C, received=0x%02X)", packetType)
			}
		}

		// Time - int64
		{
			var time int64

			if err := binary.Read(r, binary.BigEndian, &time); err != nil {
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

	response := response.BedrockStatus{
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

	for k, value := range splitID {
		if len(strings.Trim(value, " ")) < 1 {
			continue
		}

		switch k {
		case 0:
			{
				response.Edition = pointerOf(value)

				break
			}
		case 1:
			{
				motd = value

				break
			}
		case 2:
			{
				protocolVersion, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					return nil, err
				}

				response.ProtocolVersion = &protocolVersion

				break
			}
		case 3:
			{
				response.Version = pointerOf(value)

				break
			}
		case 4:
			{
				onlinePlayers, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					return nil, err
				}

				response.OnlinePlayers = &onlinePlayers

				break
			}
		case 5:
			{
				maxPlayers, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					return nil, err
				}

				response.MaxPlayers = &maxPlayers

				break
			}
		case 6:
			{
				response.ServerID = pointerOf(value)

				break
			}
		case 7:
			{
				motd += "\n" + value

				break
			}
		case 8:
			{
				response.Gamemode = pointerOf(value)

				break
			}
		case 9:
			{
				gamemodeID, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					return nil, err
				}

				response.GamemodeID = &gamemodeID

				break
			}
		case 10:
			{
				portIPv4, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					return nil, err
				}

				portIPv4Value := uint16(portIPv4)
				response.PortIPv4 = &portIPv4Value

				break
			}
		case 11:
			{
				portIPv6, err := strconv.ParseInt(value, 10, 64)

				if err != nil {
					return nil, err
				}

				response.PortIPv6 = pointerOf(uint16(portIPv6))

				break
			}
		}
	}

	if len(motd) > 0 {
		parsedMOTD, err := formatting.Parse(motd)

		if err != nil {
			return nil, err
		}

		response.MOTD = parsedMOTD
	}

	return &response, nil
}

func parseBedrockStatusOptions(opts ...options.BedrockStatus) options.BedrockStatus {
	if len(opts) < 1 {
		options := options.BedrockStatus(defaultBedrockStatusOptions)

		options.ClientGUID = rand.Int63()

		return options
	}

	return opts[0]
}
