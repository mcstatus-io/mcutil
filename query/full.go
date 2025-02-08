package query

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/response"
)

// Full runs a query on the server and returns the full information.
func Full(ctx context.Context, hostname string, port uint16, options ...options.Query) (*response.QueryFull, error) {
	r := make(chan *response.QueryFull, 1)
	e := make(chan error, 1)

	go func() {
		result, err := performFullQuery(hostname, port, options...)

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

func performFullQuery(hostname string, port uint16, options ...options.Query) (*response.QueryFull, error) {
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

	// Full stat request packet
	// https://wiki.vg/Query#Request_3
	if err = writeFullStatRequest(conn, opts.SessionID, challengeToken); err != nil {
		return nil, err
	}

	// Full stat response packet
	// https://wiki.vg/Query#Response_3
	response, err := readFullStatResponse(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	return response, err
}

func writeFullStatRequest(w io.Writer, sessionID int32, challengeToken int32) error {
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

	// Padding - [4]byte
	if _, err := buf.Write([]byte{0x00, 0x00, 0x00, 0x00}); err != nil {
		return err
	}

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}

	return nil
}

func readFullStatResponse(r io.Reader, sessionID int32) (*response.QueryFull, error) {
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

	// Session ID - int16
	{
		var serverSessionID int32

		if err := binary.Read(r, binary.BigEndian, &serverSessionID); err != nil {
			return nil, err
		}

		if serverSessionID != sessionID {
			return nil, fmt.Errorf("query: session ID mismatch (expected=%d, received=%d)", sessionID, serverSessionID)
		}
	}

	// Padding - [11]byte
	{
		data := make([]byte, 11)

		if _, err := r.Read(data); err != nil {
			return nil, err
		}
	}

	response := response.QueryFull{
		Data:    make(map[string]string),
		Players: make([]string, 0),
	}

	// K, V section - null-terminated key,pair pair string
	{
		for {
			key, err := readNTString(r)

			if err != nil {
				return nil, err
			}

			if len(key) < 1 {
				break
			}

			value, err := readNTString(r)

			if err != nil {
				return nil, err
			}

			response.Data[key] = value
		}
	}

	// Padding - [10]byte
	{
		data := make([]byte, 10)

		if _, err := r.Read(data); err != nil {
			return nil, err
		}
	}

	// Players section - null-terminated key,value pair string
	{
		for {
			username, err := readNTString(r)

			if err != nil {
				return nil, err
			}

			if len(username) < 1 {
				break
			}

			response.Players = append(response.Players, username)
		}
	}

	return &response, nil
}

func parseQueryOptions(opts ...options.Query) options.Query {
	if len(opts) < 1 {
		options := options.Query(defaultQueryOptions)

		options.SessionID = rand.Int31() & 0x0F0F0F0F

		return options
	}

	result := opts[0]
	result.SessionID &= 0x0F0F0F0F

	return result
}
