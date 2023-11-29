package mcutil

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/v2/options"
)

// SendLegacyVote sends a legacy Votifier vote to the specified Minecraft server
func SendLegacyVote(ctx context.Context, host string, port uint16, opts options.LegacyVote) error {
	e := make(chan error, 1)

	go func() {
		e <- sendLegacyVote(host, port, opts)
	}()

	select {
	case <-ctx.Done():
		if v := ctx.Err(); v != nil {
			return v
		}

		return errors.New("context finished before server sent response")
	case v := <-e:
		return v
	}
}

func sendLegacyVote(host string, port uint16, opts options.LegacyVote) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), opts.Timeout)

	if err != nil {
		return err
	}

	defer conn.Close()

	r := bufio.NewReader(conn)

	if err = conn.SetDeadline(time.Now().Add(opts.Timeout)); err != nil {
		return err
	}

	// Handshake packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#protocol-v1-deprecated
	{
		data, err := r.ReadBytes('\n')

		if err != nil {
			return err
		}

		split := strings.Split(string(data[:len(data)-1]), " ")

		if !strings.HasPrefix(split[1], "1") && !strings.HasPrefix(split[1], "2") {
			return fmt.Errorf("vote: unknown server Votifier version: %s", split[1])
		}
	}

	// Vote packet
	// https://github.com/NuVotifier/NuVotifier/wiki/Technical-QA#protocol-v1-deprecated
	{
		block, _ := pem.Decode([]byte(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", opts.PublicKey)))

		if block == nil {
			return errors.New("invalid public key value")
		}

		key, err := x509.ParsePKIXPublicKey(block.Bytes)

		if err != nil {
			return err
		}

		publicKey, ok := key.(*rsa.PublicKey)

		if !ok {
			return fmt.Errorf("parsed invalid key type: %T", key)
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
