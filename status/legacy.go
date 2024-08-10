package status

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/mcstatus-io/mcutil/v4/formatting"
	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/response"
	"github.com/mcstatus-io/mcutil/v4/util"
)

var (
	defaultJavaStatusLegacyOptions = options.StatusLegacy{
		EnableSRV: true,
		Timeout:   time.Second * 5,
	}
)

// Legacy retrieves the status of any Java Edition Minecraft server, but with reduced properties compared to Modern().
func Legacy(ctx context.Context, hostname string, port uint16, options ...options.StatusLegacy) (*response.StatusLegacy, error) {
	r := make(chan *response.StatusLegacy, 1)
	e := make(chan error, 1)

	go func() {
		result, err := getStatusLegacy(hostname, port, options...)

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

func getStatusLegacy(hostname string, port uint16, options ...options.StatusLegacy) (*response.StatusLegacy, error) {
	var (
		opts                                   = parseJavaStatusLegacyOptions(options...)
		connectionHostname                     = hostname
		connectionPort     uint16              = port
		srvRecord          *response.SRVRecord = nil
	)

	if opts.EnableSRV && port == util.DefaultJavaPort && net.ParseIP(connectionHostname) == nil {
		record, err := util.LookupSRV(hostname)

		if err == nil && record != nil {
			connectionHostname = record.Target
			connectionPort = record.Port

			srvRecord = &response.SRVRecord{
				Host: record.Target,
				Port: record.Port,
			}
		}
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", connectionHostname, connectionPort), opts.Timeout)

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

			return &response.StatusLegacy{
				Version: &response.Version{
					Name:     *versionTree,
					Protocol: protocolVersion,
				},
				Players: response.LegacyPlayers{
					Online: onlinePlayers,
					Max:    maxPlayers,
				},
				MOTD:      *motd,
				SRVRecord: srvRecord,
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

		return &response.StatusLegacy{
			Version: nil,
			Players: response.LegacyPlayers{
				Online: onlinePlayers,
				Max:    maxPlayers,
			},
			MOTD:      *motd,
			SRVRecord: srvRecord,
		}, nil
	}
}

func parseJavaStatusLegacyOptions(opts ...options.StatusLegacy) options.StatusLegacy {
	if len(opts) < 1 {
		return defaultJavaStatusLegacyOptions
	}

	return opts[0]
}
