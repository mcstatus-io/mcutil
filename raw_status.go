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

	"github.com/mcstatus-io/mcutil/options"
)

// StatusRaw returns the raw status data of any 1.7+ Minecraft server
func StatusRaw(host string, port uint16, options ...options.JavaStatus) (map[string]interface{}, error) {
	opts := parseJavaStatusOptions(options...)

	if opts.EnableSRV && port == 25565 {
		record, err := LookupSRV(host, port)

		if err == nil && record != nil {
			host = record.Target
			port = record.Port
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

	result := make(map[string]interface{})

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

			if err = json.Unmarshal(data, &result); err != nil {
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

	return result, nil
}
