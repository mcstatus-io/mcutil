package mcutil

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
)

var (
	addressRegExp = regexp.MustCompile(`^([A-Za-z0-9.]+)(?::(\d{1,5}))?$`)
)

func writePacket(data *bytes.Buffer, w io.Writer) error {
	if _, err := writeVarInt(int32(data.Len()), w); err != nil {
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

func pointerOf[T any](v T) *T {
	return &v
}

// LookupSRV resolves any Minecraft SRV record from the DNS of the domain
func LookupSRV(host string, port uint16) (*net.SRV, error) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", host)

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
