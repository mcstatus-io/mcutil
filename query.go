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
	"time"

	"github.com/mcstatus-io/mcutil/v3/formatting"
	"github.com/mcstatus-io/mcutil/v3/options"
	"github.com/mcstatus-io/mcutil/v3/response"
)

var (
	defaultQueryOptions = options.Query{
		Timeout:   time.Second * 5,
		SessionID: 0,
	}
	magic = []byte{0xFE, 0xFD}
)

// BasicQuery runs a query on the server and returns basic information
func BasicQuery(ctx context.Context, host string, port uint16, options ...options.Query) (*response.BasicQuery, error) {
	r := make(chan *response.BasicQuery, 1)
	e := make(chan error, 1)

	go func() {
		result, err := performBasicQuery(host, port, options...)

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

// FullQuery runs a query on the server and returns the full information
func FullQuery(ctx context.Context, host string, port uint16, options ...options.Query) (*response.FullQuery, error) {
	r := make(chan *response.FullQuery, 1)
	e := make(chan error, 1)

	go func() {
		result, err := performFullQuery(host, port, options...)

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

func performBasicQuery(host string, port uint16, options ...options.Query) (*response.BasicQuery, error) {
	opts := parseQueryOptions(options...)

	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

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
	if err = writeQueryHandshakeRequestPacket(conn, opts.SessionID); err != nil {
		return nil, err
	}

	// Handshake response packet
	// https://wiki.vg/Query#Response
	challengeToken, err := readQueryHandshakeResponsePacket(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	// Basic stat request packet
	// https://wiki.vg/Query#Request_2
	if err = writeQueryBasicStatRequestPacket(conn, opts.SessionID, challengeToken); err != nil {
		return nil, err
	}

	// Basic stat response packet
	// https://wiki.vg/Query#Response_2
	response, err := readQueryBasicStatResponsePacket(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	return response, err
}

func performFullQuery(host string, port uint16, options ...options.Query) (*response.FullQuery, error) {
	opts := parseQueryOptions(options...)

	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

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
	if err = writeQueryHandshakeRequestPacket(conn, opts.SessionID); err != nil {
		return nil, err
	}

	// Handshake response packet
	// https://wiki.vg/Query#Response
	challengeToken, err := readQueryHandshakeResponsePacket(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	// Full stat request packet
	// https://wiki.vg/Query#Request_3
	if err = writeQueryFullStatRequestPacket(conn, opts.SessionID, challengeToken); err != nil {
		return nil, err
	}

	// Full stat response packet
	// https://wiki.vg/Query#Response_3
	response, err := readQueryFullStatResponsePacket(r, opts.SessionID)

	if err != nil {
		return nil, err
	}

	return response, err
}

func writeQueryHandshakeRequestPacket(w io.Writer, sessionID int32) error {
	buf := &bytes.Buffer{}

	// Magic - uint16
	if _, err := buf.Write(magic); err != nil {
		return err
	}

	// Type - byte
	if err := binary.Write(buf, binary.BigEndian, byte(0x09)); err != nil {
		return err
	}

	// Session ID - int32
	if err := binary.Write(buf, binary.BigEndian, sessionID&0x0F0F0F0F); err != nil {
		return err
	}

	if _, err := io.Copy(w, buf); err != nil {
		return err
	}

	return nil
}

func readQueryHandshakeResponsePacket(r io.Reader, sessionID int32) (int32, error) {
	// Type - byte
	{
		var packetType byte

		if err := binary.Read(r, binary.BigEndian, &packetType); err != nil {
			return 0, err
		}

		if packetType != 0x09 {
			return 0, fmt.Errorf("query: received unexpected packet type (expected=0x00, received=0x%02X)", packetType)
		}
	}

	// Session ID - int32
	{
		var serverSessionID int32

		if err := binary.Read(r, binary.BigEndian, &serverSessionID); err != nil {
			return 0, err
		}

		if serverSessionID != sessionID {
			return 0, fmt.Errorf("query: session ID mismatch (expected=%d, received=%d)", sessionID, serverSessionID)
		}
	}

	var challengeToken int32

	// Challenge Token - null-terminated string
	{
		challengeTokenString, err := readNTString(r)

		if err != nil {
			return 0, err
		}

		value, err := strconv.ParseInt(challengeTokenString, 10, 32)

		if err != nil {
			return 0, err
		}

		challengeToken = int32(value)
	}

	return challengeToken, nil
}

func writeQueryBasicStatRequestPacket(w io.Writer, sessionID int32, challengeToken int32) error {
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

func writeQueryFullStatRequestPacket(w io.Writer, sessionID int32, challengeToken int32) error {
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

func readQueryBasicStatResponsePacket(r io.Reader, sessionID int32) (*response.BasicQuery, error) {
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

	var response response.BasicQuery

	// MOTD - null-terminated string
	{
		rawMOTD, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		motd, err := formatting.Parse(rawMOTD)

		if err != nil {
			return nil, err
		}

		response.MOTD = *motd
	}

	// Game Type - null-terminated string
	{
		gameType, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		response.GameType = gameType
	}

	// Map - null-terminated string
	{
		mapName, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		response.Map = mapName
	}

	// Online Players - null-terminated string
	{
		onlinePlayersString, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		onlinePlayers, err := strconv.ParseUint(onlinePlayersString, 10, 64)

		if err != nil {
			return nil, err
		}

		response.OnlinePlayers = onlinePlayers
	}

	// Max Players - null-terminated string
	{
		maxPlayersString, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		maxPlayers, err := strconv.ParseUint(maxPlayersString, 10, 64)

		if err != nil {
			return nil, err
		}

		response.MaxPlayers = maxPlayers
	}

	// Host Port - uint16
	{
		var hostPort uint16

		if err := binary.Read(r, binary.LittleEndian, &hostPort); err != nil {
			return nil, err
		}

		response.HostPort = hostPort
	}

	// Host IP - null-terminated string
	{
		hostIP, err := readNTString(r)

		if err != nil {
			return nil, err
		}

		response.HostIP = hostIP
	}

	return &response, nil
}

func readQueryFullStatResponsePacket(r io.Reader, sessionID int32) (*response.FullQuery, error) {
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

	response := response.FullQuery{
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
