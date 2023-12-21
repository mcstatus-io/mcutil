package mcutil

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/mcstatus-io/mcutil/v3/formatting"
	"github.com/mcstatus-io/mcutil/v3/options"
	"github.com/mcstatus-io/mcutil/v3/response"
)

var (
	defaultJavaStatusLegacyOptions = options.JavaStatusLegacy{
		EnableSRV: true,
		Timeout:   time.Second * 5,
	}
)

// StatusLegacy retrieves the status of any Java Edition Minecraft server, but with reduced properties compared to Status()
func StatusLegacy(ctx context.Context, host string, port uint16, options ...options.JavaStatusLegacy) (*response.JavaStatusLegacy, error) {
	r := make(chan *response.JavaStatusLegacy, 1)
	e := make(chan error, 1)

	go func() {
		result, err := getStatusLegacy(host, port, options...)

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

func getStatusLegacy(host string, port uint16, options ...options.JavaStatusLegacy) (*response.JavaStatusLegacy, error) {
	opts := parseJavaStatusLegacyOptions(options...)

	var srvResult *response.SRVRecord = nil

	if opts.EnableSRV && port == 25565 {
		record, err := LookupSRV("tcp", host)

		if err == nil && record != nil {
			host = record.Target
			port = record.Port

			srvResult = &response.SRVRecord{
				Host: record.Target,
				Port: record.Port,
			}
		}
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	// Client to server packet
	// https://wiki.vg/Server_List_Ping#Client_to_server
	{
		if _, err = conn.Write([]byte{0xFE, 0x01}); err != nil {
			return nil, err
		}
	}

	// Server to client packet
	// https://wiki.vg/Server_List_Ping#Server_to_client
	{
		// Packet Type - byte
		{
			var packetType byte

			if err := binary.Read(conn, binary.BigEndian, &packetType); err != nil {
				return nil, err
			}

			if packetType != 0xFF {
				return nil, fmt.Errorf("status: received unexpected packet type (expected=0xFF, received=0x%02X)", packetType)
			}
		}

		var packetLength uint16

		// Length - uint16
		{
			if err = binary.Read(conn, binary.BigEndian, &packetLength); err != nil {
				return nil, err
			}

			if packetLength < 2 {
				return nil, fmt.Errorf("status: received status response with no data (bytes=%d)", packetLength)
			}
		}

		var data []uint16

		// Packet Data
		{
			data = make([]uint16, packetLength)

			if err = binary.Read(conn, binary.BigEndian, &data); err != nil {
				return nil, err
			}
		}

		result := string(utf16.Decode(data))

		// TODO clean up this code at some point, avoid using string functions

		if data[0] == 0x00A7 && data[1] == 0x0031 {
			// 1.4+ server

			split := strings.Split(result, "\x00")

			if len(split) < 6 {
				return nil, fmt.Errorf("status: not enough information received (expected=6, received=%d)", len(split))
			}

			protocolVersion, err := strconv.ParseInt(split[1], 10, 32)

			if err != nil {
				return nil, err
			}

			versionTree, err := formatting.Parse(split[2])

			if err != nil {
				return nil, err
			}

			motd, err := formatting.Parse(split[3])

			if err != nil {
				return nil, err
			}

			onlinePlayers, err := strconv.ParseInt(split[4], 10, 32)

			if err != nil {
				return nil, err
			}

			maxPlayers, err := strconv.ParseInt(split[5], 10, 32)

			if err != nil {
				return nil, err
			}

			return &response.JavaStatusLegacy{
				Version: &response.Version{
					NameRaw:   versionTree.Raw,
					NameClean: versionTree.Clean,
					NameHTML:  versionTree.HTML,
					Protocol:  protocolVersion,
				},
				Players: response.LegacyPlayers{
					Online: onlinePlayers,
					Max:    maxPlayers,
				},
				MOTD:      *motd,
				SRVResult: srvResult,
			}, nil
		}

		// < 1.4 server

		split := strings.Split(result, "\u00A7")

		if len(split) < 3 {
			return nil, fmt.Errorf("status: not enough information received (expected=3, received=%d)", len(split))
		}

		motd, err := formatting.Parse(split[0])

		if err != nil {
			return nil, err
		}

		onlinePlayers, err := strconv.ParseInt(split[1], 10, 32)

		if err != nil {
			return nil, err
		}

		maxPlayers, err := strconv.ParseInt(split[2], 10, 32)

		if err != nil {
			return nil, err
		}

		return &response.JavaStatusLegacy{
			Version: nil,
			Players: response.LegacyPlayers{
				Online: onlinePlayers,
				Max:    maxPlayers,
			},
			MOTD:      *motd,
			SRVResult: srvResult,
		}, nil
	}
}

func parseJavaStatusLegacyOptions(opts ...options.JavaStatusLegacy) options.JavaStatusLegacy {
	if len(opts) < 1 {
		return defaultJavaStatusLegacyOptions
	}

	return opts[0]
}
