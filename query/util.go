package query

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mcstatus-io/mcutil/v4/options"
)

var (
	defaultQueryOptions = options.Query{
		Timeout:   time.Second * 5,
		SessionID: 0,
	}
	magic = []byte{0xFE, 0xFD}
)

func convertISO8859ToUTF8(data []byte) string {
	result := make([]rune, len(data))

	for i, b := range data {
		result[i] = rune(b)
	}

	return string(result)
}

func readNTString(r io.Reader) (string, error) {
	data := make([]byte, 0)

	for {
		chunk := make([]byte, 1)

		if _, err := r.Read(chunk); err != nil {
			return "", err
		}

		if chunk[0] == 0x00 {
			break
		}

		data = append(data, chunk...)
	}

	return convertISO8859ToUTF8(data), nil
}

func writeHandshakeRequest(w io.Writer, sessionID int32) error {
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

func readHandshakeResponse(r io.Reader, sessionID int32) (int32, error) {
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
