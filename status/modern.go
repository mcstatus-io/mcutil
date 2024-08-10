package status

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/v4/formatting"
	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/proto"
	"github.com/mcstatus-io/mcutil/v4/response"
	"github.com/mcstatus-io/mcutil/v4/util"
)

var defaultJavaStatusOptions = options.StatusModern{
	EnableSRV:       true,
	Timeout:         time.Second * 5,
	ProtocolVersion: -1,
	Ping:            true,
	Debug:           false,
}

type rawJavaStatus struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int64  `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    *int64 `json:"max"`
		Online *int64 `json:"online"`
		Sample []struct {
			ID   interface{} `json:"id"`
			Name string      `json:"name"`
		} `json:"sample"`
	} `json:"players"`
	Description interface{} `json:"description"`
	Favicon     *string     `json:"favicon"`
	ModInfo     struct {
		List []struct {
			ID      string `json:"modid"`
			Version string `json:"version"`
		} `json:"modList"`
		Type string `json:"type"`
	} `json:"modinfo"`
	ForgeData struct {
		Channels []struct {
			Required bool   `json:"required"`
			Res      string `json:"res"`
			Version  string `json:"version"`
		} `json:"channels"`
		FMLNetworkVersion int `json:"fmlNetworkVersion"`
		Mods              []struct {
			ID      string `json:"modId"`
			Version string `json:"modmarker"`
		} `json:"mods"`
	} `json:"forgeData"`
}

// Modern retrieves the status of any 1.7+ Minecraft server.
func Modern(ctx context.Context, hostname string, port uint16, options ...options.StatusModern) (*response.StatusModern, error) {
	r := make(chan *response.StatusModern, 1)
	e := make(chan error, 1)

	go func() {
		result, err := getStatusModern(hostname, port, options...)

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

func getStatusModern(hostname string, port uint16, options ...options.StatusModern) (*response.StatusModern, error) {
	var (
		opts                                   = parseJavaStatusOptions(options...)
		connectionHostname string              = hostname
		connectionPort     uint16              = port
		srvRecord          *response.SRVRecord = nil
		rawResponse        rawJavaStatus       = rawJavaStatus{}
		latency            time.Duration       = 0
	)

	if opts.EnableSRV && port == util.DefaultJavaPort && net.ParseIP(connectionHostname) == nil {
		record, err := util.LookupSRV(hostname)

		if err == nil && record != nil {
			connectionHostname = record.Target
			connectionPort = record.Port

			srvRecord = &response.SRVRecord{
				Host: record.Target,
				Port: record.Port,
			}

			if opts.Debug {
				log.Printf("Found an SRV record (host=%s, port=%d)", record.Target, record.Port)
			}
		} else if opts.Debug {
			log.Println("Could not find an SRV record for this host")
		}
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", connectionHostname, connectionPort), opts.Timeout)

	if err != nil {
		return nil, err
	}

	if opts.Debug {
		log.Printf("Successfully connected to %s:%d\n", connectionHostname, connectionPort)
	}

	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return nil, err
	}

	if err = writeJavaStatusHandshakePacket(conn, int32(opts.ProtocolVersion), hostname, port); err != nil {
		return nil, err
	}

	if opts.Debug {
		log.Printf("[S <- C] Wrote handshake packet (proto=%d, host=%s, port=%d, next_state=0)\n", opts.ProtocolVersion, hostname, port)
	}

	if err = writeJavaStatusStatusRequestPacket(conn); err != nil {
		return nil, err
	}

	if opts.Debug {
		log.Println("[S <- C] Wrote status request packet")
	}

	if err = readJavaStatusStatusResponsePacket(conn, &rawResponse); err != nil {
		return nil, err
	}

	if opts.Debug {
		log.Println("[S -> C] Read status response packet")
	}

	if opts.Ping {
		payload := rand.Int63()

		if err = writeJavaStatusPingPacket(conn, payload); err != nil {
			return nil, err
		}

		if opts.Debug {
			log.Printf("[S <- C] Wrote ping packet (payload=%d)\n", payload)
		}

		pingStart := time.Now()

		if err = readJavaStatusPongPacket(conn, payload); err != nil {
			return nil, err
		}

		if opts.Debug {
			log.Printf("[S -> C] Read ping packet (payload=%d)\n", payload)
		}

		latency = time.Since(pingStart)
	}

	return formatJavaStatusResponse(rawResponse, srvRecord, latency)
}

func parseJavaStatusOptions(opts ...options.StatusModern) options.StatusModern {
	if len(opts) < 1 {
		return defaultJavaStatusOptions
	}

	return opts[0]
}

// https://wiki.vg/Server_List_Ping#Handshake
func writeJavaStatusHandshakePacket(w io.Writer, protocolVersion int32, host string, port uint16) error {
	buf := &bytes.Buffer{}

	// Packet ID - varint
	if err := proto.WriteVarInt(0x00, buf); err != nil {
		return err
	}

	// Protocol version - varint
	if err := proto.WriteVarInt(protocolVersion, buf); err != nil {
		return err
	}

	// Host - string
	if err := proto.WriteString(host, buf); err != nil {
		return err
	}

	// Port - uint16
	if err := binary.Write(buf, binary.BigEndian, port); err != nil {
		return err
	}

	// Next state - varint
	if err := proto.WriteVarInt(1, buf); err != nil {
		return err
	}

	return writePacket(w, buf)
}

// https://wiki.vg/Server_List_Ping#Request
func writeJavaStatusStatusRequestPacket(w io.Writer) error {
	buf := &bytes.Buffer{}

	// Packet ID - varint
	if err := proto.WriteVarInt(0x00, buf); err != nil {
		return err
	}

	return writePacket(w, buf)
}

// https://wiki.vg/Server_List_Ping#Response
func readJavaStatusStatusResponsePacket(r io.Reader, result interface{}) error {
	// Packet length - varint
	{
		if _, err := proto.ReadVarInt(r); err != nil {
			return err
		}
	}

	// Packet type - varint
	{
		packetType, err := proto.ReadVarInt(r)

		if err != nil {
			return err
		}

		if packetType != 0x00 {
			return fmt.Errorf("status: received unexpected packet type (expected=0x00, received=0x%02X)", packetType)
		}
	}

	// Data - string
	{
		data, err := proto.ReadString(r)

		if err != nil {
			return err
		}

		if err = json.Unmarshal(data, result); err != nil {
			return err
		}
	}

	return nil
}

// https://wiki.vg/Server_List_Ping#Ping
func writeJavaStatusPingPacket(w io.Writer, payload int64) error {
	buf := &bytes.Buffer{}

	// Packet ID - varint
	if err := proto.WriteVarInt(0x01, buf); err != nil {
		return err
	}

	// Payload - int64
	if err := binary.Write(buf, binary.BigEndian, payload); err != nil {
		return err
	}

	return writePacket(w, buf)
}

// https://wiki.vg/Server_List_Ping#Pong
func readJavaStatusPongPacket(r io.Reader, payload int64) error {
	// Packet length - varint
	{
		if _, err := proto.ReadVarInt(r); err != nil {
			return err
		}
	}

	// Packet type - varint
	{
		packetType, err := proto.ReadVarInt(r)

		if err != nil {
			return err
		}

		if packetType != 0x01 {
			return fmt.Errorf("status: received unexpected packet type (expected=0x01, received=0x%02X)", packetType)
		}
	}

	// Payload - int64
	{
		var returnPayload int64

		if err := binary.Read(r, binary.BigEndian, &returnPayload); err != nil {
			return err
		}

		if payload != returnPayload {
			return fmt.Errorf("status: received unexpected payload (expected=%X, received=%x)", payload, returnPayload)
		}
	}

	return nil
}

func formatJavaStatusResponse(serverResponse rawJavaStatus, srvRecord *response.SRVRecord, latency time.Duration) (*response.StatusModern, error) {
	motd, err := formatting.Parse(serverResponse.Description)

	if err != nil {
		return nil, err
	}

	samplePlayers := make([]response.SamplePlayer, 0)

	if serverResponse.Players.Sample != nil {
		for _, player := range serverResponse.Players.Sample {
			name, err := formatting.Parse(player.Name)

			if err != nil {
				return nil, err
			}

			uuid, ok := parsePlayerID(player.ID)

			if !ok {
				return nil, fmt.Errorf("status: invalid player UUID: %+v", player.ID)
			}

			samplePlayers = append(samplePlayers, response.SamplePlayer{
				ID:   uuid,
				Name: *name,
			})
		}
	}

	version, err := formatting.Parse(serverResponse.Version.Name)

	if err != nil {
		return nil, err
	}

	result := &response.StatusModern{
		Version: response.Version{
			Name:     *version,
			Protocol: serverResponse.Version.Protocol,
		},
		Players: response.Players{
			Online: serverResponse.Players.Online,
			Max:    serverResponse.Players.Max,
			Sample: samplePlayers,
		},
		MOTD:      *motd,
		Favicon:   serverResponse.Favicon,
		SRVRecord: srvRecord,
		Latency:   latency,
		Mods:      nil,
	}

	if len(serverResponse.ModInfo.Type) > 0 {
		mods := make([]response.Mod, 0)

		for _, mod := range serverResponse.ModInfo.List {
			mods = append(mods, response.Mod{
				ID:      mod.ID,
				Version: mod.Version,
			})
		}

		result.Mods = &response.ModInfo{
			Type: serverResponse.ModInfo.Type,
			List: mods,
		}
	}

	if serverResponse.ForgeData.Mods != nil {
		mods := make([]response.Mod, 0)

		for _, mod := range serverResponse.ForgeData.Mods {
			mods = append(mods, response.Mod{
				ID:      mod.ID,
				Version: mod.Version,
			})
		}

		result.Mods = &response.ModInfo{
			Type: "FML2",
			List: mods,
		}
	}

	return result, nil
}
