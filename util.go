package mcutil

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
)

const (
	DefaultJavaPort    = 25565
	DefaultBedrockPort = 19132
)

var (
	addressRegExp = regexp.MustCompile(`^([A-Za-z0-9.]+)(?::(\d{1,5}))?$`)
)

func pointerOf[T any](v T) *T {
	return &v
}

func writePacket(w io.Writer, data *bytes.Buffer) error {
	if err := writeVarInt(int32(data.Len()), w); err != nil {
		return err
	}

	_, err := io.Copy(w, data)

	return err
}

func readNTString(r io.Reader) (string, error) {
	result := make([]byte, 0)

	for {
		data := make([]byte, 1)

		if _, err := r.Read(data); err != nil {
			return "", err
		}

		if data[0] == 0x00 {
			break
		}

		result = append(result, data...)
	}

	return string(result), nil
}

func parsePlayerID(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []interface{}:
		{
			if len(v) != 4 {
				return "", false
			}

			var a, b uint64

			for i, val := range v {
				parsed, ok := val.(float64)

				if !ok {
					return "", false
				}

				if i < 2 {
					a |= uint64(parsed) << ((i % 2) * 8)
				} else {
					b |= uint64(parsed) << ((i % 2) * 8)
				}
			}

			return fmt.Sprintf("%016x%016x", a, b), true
		}
	default:
		return "", false
	}
}

// LookupSRV resolves any Minecraft SRV record from the DNS of the domain
func LookupSRV(protocol, host string) (*net.SRV, error) {
	_, addrs, err := net.LookupSRV("minecraft", protocol, host)

	if err != nil {
		return nil, err
	}

	if len(addrs) < 1 {
		return nil, nil
	}

	return addrs[0], nil
}

// ParseAddress parses the host and port out of an address string
func ParseAddress(address string, defaultPort uint16) (string, uint16, error) {
	matches := addressRegExp.FindAllStringSubmatch(address, -1)

	if matches == nil || len(matches) < 1 {
		return "", defaultPort, fmt.Errorf("address: cannot parse \"%s\"", address)
	}

	if len(matches[0]) < 3 || len(matches[0][2]) < 1 {
		return matches[0][1], defaultPort, nil
	}

	port, err := strconv.ParseUint(matches[0][2], 10, 16)

	if err != nil {
		return "", defaultPort, err
	}

	return matches[0][1], uint16(port), nil
}
