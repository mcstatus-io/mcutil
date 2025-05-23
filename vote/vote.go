package vote

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/v4/options"
)

var (
	// ErrPublicKeyRequired means that the server is using Votifier 1 but the PublicKey option is missing.
	ErrPublicKeyRequired = errors.New("vote: server negotiated Votifier 1, but PublicKey option is empty")
	// ErrInvalidPublicKey means the public key provided cannot be parsed.
	ErrInvalidPublicKey = errors.New("vote: invalid public key value")
	// ErrPublicKeyRequired means that the server is using Votifier 2 but the Token option is missing.
	ErrTokenRequired = errors.New("vote: server negotiated Votifier 2, but Token option is empty")
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

// SendVote sends a Votifier vote to the specified Minecraft server.
func SendVote(ctx context.Context, host string, port uint16, opts options.Vote) error {
	e := make(chan error, 1)

	go func() {
		e <- sendVote(host, port, opts)
	}()

	select {
	case <-ctx.Done():
		if v := ctx.Err(); v != nil {
			return v
		}

		return context.DeadlineExceeded
	case v := <-e:
		return v
	}
}

func sendVote(host string, port uint16, opts options.Vote) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return err
	}

	var (
		challenge    string
		version      string
		majorVersion string
	)

	// Handshake packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#handshake
	{
		data, err := r.ReadBytes('\n')

		if err != nil {
			return err
		}

		dataSegments := strings.Split(string(data[:len(data)-1]), " ")

		if len(dataSegments) < 2 {
			return fmt.Errorf("vote: server sent invalid handshake packet with value: %s", data[:len(data)-1])
		}

		version = dataSegments[1]
		majorVersion = strings.Split(version, ".")[0]

		if len(dataSegments) > 2 {
			challenge = dataSegments[2]
		}
	}

	if majorVersion != "2" && majorVersion != "1" {
		return fmt.Errorf("vote: unknown Votifier version: %s", version)
	}

	if majorVersion == "2" && len(opts.Token) > 0 {
		if err := sendVotifier2Vote(r, conn, host, port, challenge, opts); err != nil {
			return err
		}
	} else if len(opts.PublicKey) > 0 {
		if err := sendVotifier1Vote(conn, opts); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("vote: version mismatch (server is expecting a Votifier %s packet, but options are missing to allow sending of this vote packet or any prior backwards-compatible versions)", version)
	}

	return nil
}

func sendVotifier1Vote(conn net.Conn, opts options.Vote) error {
	if len(opts.IPAddress) < 1 {
		opts.IPAddress = "127.0.0.1"
	}

	// Vote packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#protocol-v1-deprecated
	{
		block, _ := pem.Decode([]byte(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", opts.PublicKey)))

		if block == nil {
			return ErrInvalidPublicKey
		}

		key, err := x509.ParsePKIXPublicKey(block.Bytes)

		if err != nil {
			return err
		}

		publicKey, ok := key.(*rsa.PublicKey)

		if !ok {
			return fmt.Errorf("vote: parsed invalid key type: %T", key)
		}

		payload := fmt.Sprintf(
			"VOTE\n%s\n%s\n%s\n%s",
			opts.ServiceName,
			opts.Username,
			opts.IPAddress,
			opts.Timestamp.Format(time.RFC3339),
		)

		encryptedPayload, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(payload))

		if err != nil {
			return err
		}

		if _, err = conn.Write(encryptedPayload); err != nil {
			return err
		}
	}

	return nil
}

func sendVotifier2Vote(r *bufio.Reader, conn net.Conn, host string, port uint16, challenge string, opts options.Vote) error {
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
			break
		case "error":
			return fmt.Errorf("vote: server returned error: %s", response.Error)
		default:
			return fmt.Errorf("vote: received unexpected server response (expected=<nil>, received=%s)", response.Status)
		}
	}

	return nil
}
