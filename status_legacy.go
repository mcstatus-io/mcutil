package mcstatus

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/mcstatus-io/mcutil/description"
	"github.com/mcstatus-io/mcutil/options"
	"github.com/mcstatus-io/mcutil/response"
)

var (
	defaultJavaStatusLegacyOptions = options.JavaStatusLegacy{
		EnableSRV: true,
		Timeout:   time.Second * 5,
	}
)

func StatusLegacy(host string, port uint16, options ...options.JavaStatusLegacy) (*response.JavaStatusLegacy, error) {
	opts := parseJavaStatusLegacyOptions(options...)

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

	conn, err := net.DialTimeout("tcp4", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

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
		packetType, err := r.ReadByte()

		if err != nil {
			return nil, err
		}

		if packetType != 0xFF {
			return nil, fmt.Errorf("unexpected packet type returned from server: 0x%X", packetType)
		}

		var length uint16

		if err = binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}

		data := make([]byte, length*2)

		if _, err = r.Read(data); err != nil {
			return nil, err
		}

		byteData := make([]uint16, length)

		for i, l := 0, len(data); i < l; i += 2 {
			byteData[i/2] = (uint16(data[i]) << 8) | uint16(data[i+1])
		}

		result := string(utf16.Decode(byteData))

		if byteData[0] == 0x00A7 && byteData[1] == 0x0031 {
			// 1.4+ server

			split := strings.Split(result, "\x00")

			protocolVersion, err := strconv.ParseInt(split[1], 10, 32)

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

			motd, err := description.ParseMOTD(split[3])

			if err != nil {
				return nil, err
			}

			versionTree, err := description.ParseMOTD(split[2])

			if err != nil {
				return nil, err
			}

			return &response.JavaStatusLegacy{
				Version: &response.Version{
					NameRaw:   versionTree.Raw,
					NameClean: versionTree.Clean,
					NameHTML:  versionTree.HTML,
					Protocol:  int(protocolVersion),
				},
				Players: response.LegacyPlayers{
					Online: int(onlinePlayers),
					Max:    int(maxPlayers),
				},
				MOTD:      *motd,
				SRVResult: srvResult,
			}, nil
		} else {
			// < 1.4 server

			split := strings.Split(result, "\u00A7")

			onlinePlayers, err := strconv.ParseInt(split[1], 10, 32)

			if err != nil {
				return nil, err
			}

			maxPlayers, err := strconv.ParseInt(split[2], 10, 32)

			if err != nil {
				return nil, err
			}

			motd, err := description.ParseMOTD(split[0])

			if err != nil {
				return nil, err
			}

			return &response.JavaStatusLegacy{
				Version: nil,
				Players: response.LegacyPlayers{
					Online: int(onlinePlayers),
					Max:    int(maxPlayers),
				},
				MOTD:      *motd,
				SRVResult: srvResult,
			}, nil
		}
	}
}

func parseJavaStatusLegacyOptions(opts ...options.JavaStatusLegacy) options.JavaStatusLegacy {
	if len(opts) < 1 {
		return defaultJavaStatusLegacyOptions
	}

	return opts[0]
}
