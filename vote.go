package mcutil

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/options"
)

type voteMessage struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

type votePayload struct {
	ServiceName string `json:"serviceName"`
	Username    string `json:"username"`
	Address     string `json:"address"`
	Timestamp   int64  `json:"timestamp"`
	Challenge   string `json:"challenge"`
	UUID        string `json:"uuid,omitempty"`
}

type voteResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// SendVote sends a Votifier vote to the specified Minecraft server
func SendVote(host string, port uint16, opts options.Vote) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return err
	}

	var challenge string

	// Handshake packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#handshake
	{
		data, err := r.ReadBytes('\n')

		if err != nil {
			return err
		}

		split := strings.Split(string(data[:len(data)-1]), " ")

		if split[1] != "2" {
			return fmt.Errorf("vote: unknown server Votifier version: %s", split[1])
		}

		challenge = split[2]
	}

	// Vote packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#protocol-v2
	{
		buf := &bytes.Buffer{}

		payload := votePayload{
			ServiceName: opts.ServiceName,
			Username:    opts.Username,
			Address:     fmt.Sprintf("%s:%d", host, port),
			Timestamp:   opts.Timestamp.UnixMilli(),
			Challenge:   challenge,
			UUID:        opts.UUID,
		}

		payloadData, err := json.Marshal(payload)

		if err != nil {
			return err
		}

		hash := hmac.New(sha256.New, []byte(opts.Token))
		hash.Write(payloadData)

		message := voteMessage{
			Payload:   string(payloadData),
			Signature: base64.StdEncoding.EncodeToString(hash.Sum(nil)),
		}

		messageData, err := json.Marshal(message)

		if err != nil {
			return err
		}

		if err := binary.Write(buf, binary.BigEndian, uint16(0x733A)); err != nil {
			return err
		}

		if err := binary.Write(buf, binary.BigEndian, uint16(len(messageData))); err != nil {
			return err
		}

		if _, err := buf.Write(messageData); err != nil {
			return err
		}

		if _, err := io.Copy(conn, buf); err != nil {
			return err
		}
	}

	// Response packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#protocol-v2
	{
		data, err := r.ReadBytes('\n')

		if err != nil {
			return err
		}

		response := voteResponse{}

		if err = json.Unmarshal(data[:len(data)-1], &response); err != nil {
			return err
		}

		switch response.Status {
		case "ok":
			{
				return nil
			}
		case "error":
			{
				return fmt.Errorf("server returned error: %s", response.Error)
			}
		default:
			{
				return fmt.Errorf("vote: received unexpected server response (expected=ok, received=%s)", response.Status)
			}
		}
	}
}
