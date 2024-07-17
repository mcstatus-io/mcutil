package util

import (
	"net"
	"strconv"
)

const (
	// DefaultJavaPort is the default port used by Minecraft Java Edition servers.
	DefaultJavaPort = 25565
	// DefaultBedrockPort is the default port used by Minecraft Bedrock Edition servers.
	DefaultBedrockPort = 19132
)

// LookupSRV resolves any Minecraft SRV record from the DNS of the domain.
func LookupSRV(host string) (*net.SRV, error) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", host)

	if err != nil {
		return nil, err
	}

	if len(addrs) < 1 {
		return nil, nil
	}

	return addrs[0], nil
}

// ParseAddress parses the host and port out of an address string. This method will return a nil
// port if there is not one specified in the string.
func ParseAddress(host string) (string, *uint16, error) {
	hostname, port, err := net.SplitHostPort(host)

	if err != nil {
		addrError, ok := err.(*net.AddrError)

		if !ok {
			return "", nil, err
		}

		if addrError.Err == "missing port in address" {
			return host, nil, nil
		}
	}

	parsedPort, err := strconv.ParseUint(port, 10, 16)

	if err != nil {
		return "", nil, err
	}

	newPort := uint16(parsedPort)

	return hostname, &newPort, nil
}
