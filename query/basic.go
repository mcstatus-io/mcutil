package query

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/mcstatus-io/mcutil/v4/formatting"
	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/response"
)

// Basic runs a query on the server and returns basic information.
func Basic(ctx context.Context, hostname string, port uint16, options ...options.Query) (*response.QueryBasic, error) {
	r := make(chan *response.QueryBasic, 1)
	e := make(chan error, 1)

	go func() {
		result, err := performBasicQuery(hostname, port, options...)

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

func performBasicQuery(hostname string, port uint16, options ...options.Query) (*response.QueryBasic, error) {
	opts := parseQueryOptions(options...)

	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", hostname, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	// Handshake request packet
	// https://wiki.vg/Query#Request
	if err = writeHandshakeRequest(conn, opts.SessionID); err != nil {
		return nil, err
	}

	// Handshake response packet
	// https://wiki.vg/Query#Response
	challengeToken, err := readHandshakeResponse(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	// Basic stat request packet
	// https://wiki.vg/Query#Request_2
	if err = writeBasicStatRequest(conn, opts.SessionID, challengeToken); err != nil {
		return nil, err
	}

	// Basic stat response packet
	// https://wiki.vg/Query#Response_2
	response, err := readBasicStatResponse(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	return response, err
}

func writeBasicStatRequest(w io.Writer, sessionID int32, challengeToken int32) error {
	buf := &bytes.Buffer{}

	// Magic - uint16
	if _, err := buf.Write(magic); err != nil {
		return err
	}

	// Type - byte
	if err := binary.Write(buf, binary.BigEndian, byte(0x00)); err != nil {
		return err
	}

	// Session ID - int32
	if err := binary.Write(buf, binary.BigEndian, sessionID&0x0F0F0F0F); err != nil {
		return err
	}

	// Challenge Token - int32
	if err := binary.Write(buf, binary.BigEndian, challengeToken); err != nil {
		return err
	}

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}

	return nil
}

func readBasicStatResponse(r io.Reader, sessionID int32) (*response.QueryBasic, error) {
	// Type - byte
	{
		var packetType byte

		if err := binary.Read(r, binary.BigEndian, &packetType); err != nil {
			return nil, err
		}

		if packetType != 0x00 {
			return nil, fmt.Errorf("query: received unexpected packet type (expected=0x00, received=0x%02X)", packetType)
		}
	}

	// Session ID - int32
	{
		var serverSessionID int32

		if err := binary.Read(r, binary.BigEndian, &serverSessionID); err != nil {
			return nil, err
		}

		if serverSessionID != sessionID {
			return nil, fmt.Errorf("query: session ID mismatch (expected=%d, received=%d)", sessionID, serverSessionID)
		}
	}

	var response response.QueryBasic

	// MOTD - null-terminated string
	{
		value, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		motd, err := formatting.Parse(value)

		if err != nil {
			return nil, err
		}

		response.MOTD = *motd
	}

	// Game Type - null-terminated string
	{
		value, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		response.GameType = value
	}

	// Map - null-terminated string
	{
		value, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		response.Map = value
	}

	// Online Players - null-terminated string
	{
		value, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		onlinePlayers, err := strconv.ParseUint(value, 10, 64)

		if err != nil {
			return nil, err
		}

		response.OnlinePlayers = onlinePlayers
	}

	// Max Players - null-terminated string
	{
		value, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		maxPlayers, err := strconv.ParseUint(value, 10, 64)

		if err != nil {
			return nil, err
		}

		response.MaxPlayers = maxPlayers
	}

	// Host Port - uint16
	{
		var value uint16

		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return nil, err
		}

		response.HostPort = value
	}

	// Host IP - null-terminated string
	{
		value, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		response.HostIP = value
	}

	return &response, nil
}
