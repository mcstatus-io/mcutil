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

func decodeASCII(input []byte) string {
	data := make([]rune, len(input))

	for i, b := range input {
		data[i] = rune(b)
	}

	return string(data)
}

func writePacket(data *bytes.Buffer, w io.Writer) error {
	if _, err := writeVarInt(int32(data.Len()), w); err != nil {
		return err
	}

	_, err := io.Copy(w, data)

	return err
}

// Resolves any Minecraft SRV record from the DNS of the website
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
		return "", defaultPort, fmt.Errorf("address \"%s\" does not match any known format", address)
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
