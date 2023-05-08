package mcutil

import (
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

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	if err = writeJavaStatusHandshakeRequestPacket(conn, int32(opts.ProtocolVersion), host, port); err != nil {
		return nil, err
	}

	if err = writeJavaStatusStatusRequestPacket(conn); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	if err = readJavaStatusStatusResponsePacket(conn, &result); err != nil {
		return nil, err
	}

	payload := rand.Int63()

	if err = writeJavaStatusPingPacket(conn, payload); err != nil {
		return nil, err
	}

	if err = readJavaStatusPongPacket(conn, payload); err != nil {
		return nil, err
	}

	return result, nil
}
