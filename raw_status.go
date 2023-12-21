package mcutil

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/v3/options"
)

// StatusRaw returns the raw status data of any 1.7+ Minecraft server
func StatusRaw(ctx context.Context, host string, port uint16, options ...options.JavaStatus) (map[string]interface{}, error) {
	r := make(chan map[string]interface{}, 1)
	e := make(chan error, 1)

	go func() {
		result, err := getStatusRaw(host, port, options...)

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

func getStatusRaw(host string, port uint16, options ...options.JavaStatus) (map[string]interface{}, error) {
	opts := parseJavaStatusOptions(options...)

	if opts.EnableSRV && port == 25565 {
		record, err := LookupSRV("tcp", host)

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
