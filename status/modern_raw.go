package status

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/util"
)

// ModernRaw returns the raw status data of any 1.7+ Java Edition Minecraft server.
func ModernRaw(ctx context.Context, hostname string, port uint16, options ...options.StatusModern) (map[string]interface{}, error) {
	r := make(chan map[string]interface{}, 1)
	e := make(chan error, 1)

	go func() {
		result, err := getStatusRaw(hostname, port, options...)

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

func getStatusRaw(hostname string, port uint16, options ...options.StatusModern) (map[string]interface{}, error) {
	var (
		opts                                      = parseJavaStatusOptions(options...)
		connectionHostname                        = hostname
		connectionPort     uint16                 = port
		result             map[string]interface{} = make(map[string]interface{})
		payload            int64                  = rand.Int63()
	)

	if opts.EnableSRV && port == util.DefaultJavaPort && net.ParseIP(connectionHostname) == nil {
		record, err := util.LookupSRV(hostname)

		if err == nil && record != nil {
			connectionHostname = record.Target
			connectionPort = record.Port
		}
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", connectionHostname, connectionPort), opts.Timeout)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	if err = writeJavaStatusHandshakePacket(conn, int32(opts.ProtocolVersion), connectionHostname, connectionPort); err != nil {
		return nil, err
	}

	if err = writeJavaStatusStatusRequestPacket(conn); err != nil {
		return nil, err
	}

	if err = readJavaStatusStatusResponsePacket(conn, &result); err != nil {
		return nil, err
	}

	if err = writeJavaStatusPingPacket(conn, payload); err != nil {
		return nil, err
	}

	if err = readJavaStatusPongPacket(conn, payload); err != nil {
		return nil, err
	}

	return result, nil
}
