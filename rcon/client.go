package rcon

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/mcstatus-io/mcutil/v4/options"
)

var (
	// ErrNotConnected means the client attempted to send data but there was no connection to the server.
	ErrNotConnected = errors.New("rcon: not connected to the server")
	// ErrAlreadyLoggedIn means the RCON client was already logged in but a second login attempt was made.
	ErrAlreadyLoggedIn = errors.New("rcon: already successfully logged in")
	// ErrInvalidPassword means the password used in the RCON login was incorrect.
	ErrInvalidPassword = errors.New("rcon: incorrect password")
	// ErrNotAuthenticated means the client attempted to execute a command before a login was successful.
	ErrNotAuthenticated = errors.New("rcon: not authenticated with the server")
)

var (
	defaultOptions = options.RCON{
		Timeout: time.Second * 5,
	}
)

// Client is a client for interacting with RCON and contains multiple methods.
type Client struct {
	conn        net.Conn
	Messages    chan string
	runTrigger  chan bool
	authSuccess bool
	requestID   int32
}

// Dial connects to the server using the address provided and returns a new client.
func Dial(hostname string, port uint16, options ...options.RCON) (*Client, error) {
	opts := parseOptions(options...)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), opts.Timeout)

	if err != nil {
		return nil, err
	}

	return &Client{
		conn:        conn,
		Messages:    make(chan string),
		runTrigger:  make(chan bool),
		authSuccess: false,
		requestID:   0,
	}, nil
}

// Login communicates authentication with the server using the plaintext password.
func (r *Client) Login(password string) error {
	if r.conn == nil {
		return ErrNotConnected
	}

	if r.authSuccess {
		return ErrAlreadyLoggedIn
	}

	requestID := r.nextRequestID()

	// Login request packet
	// https://wiki.vg/RCON#3:_Login
	{
		buf := &bytes.Buffer{}

		// Length - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(10+len(password))); err != nil {
			return err
		}

		// Request ID - int32
		if err := binary.Write(buf, binary.LittleEndian, requestID); err != nil {
			return err
		}

		// Type - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(3)); err != nil {
			return err
		}

		// Payload - null-terminated string
		if _, err := buf.Write(append([]byte(password), 0x00)); err != nil {
			return err
		}

		// Padding - byte
		if err := buf.WriteByte(0x00); err != nil {
			return err
		}

		if _, err := io.Copy(r.conn, buf); err != nil {
			return err
		}
	}

	// Login response packet
	// https://wiki.vg/RCON#3:_Login
	{
		var packetLength int32

		// Length - int32
		{
			if err := binary.Read(r.conn, binary.LittleEndian, &packetLength); err != nil {
				return err
			}
		}

		// Request ID - int32
		{
			var reqID int32

			if err := binary.Read(r.conn, binary.LittleEndian, &reqID); err != nil {
				return err
			}

			if reqID == -1 {
				return ErrInvalidPassword
			} else if reqID != requestID {
				return fmt.Errorf("rcon: received unexpected request ID (expected=%d, received=%d)", requestID, reqID)
			}
		}

		// Type - int32
		{
			var packetType int32

			if err := binary.Read(r.conn, binary.LittleEndian, &packetType); err != nil {
				return err
			}

			if packetType != 0x02 {
				return fmt.Errorf("rcon: received unexpected packet type (expected=0x02, received=0x%02X)", packetType)
			}
		}

		// Remaining bytes
		{
			data := make([]byte, packetLength-8)

			if _, err := r.conn.Read(data); err != nil {
				return err
			}
		}
	}

	r.authSuccess = true

	go (func() {
		for {
			err := r.readMessage()

			if err != nil {
				fmt.Println(err)

				r.conn.Close()
				r.conn = nil

				break
			}
		}
	})()

	return nil
}

// Run executes the command on the server but does not wait for a response.
func (r *Client) Run(command string) error {
	if r.conn == nil {
		return ErrNotConnected
	}

	if !r.authSuccess {
		return ErrNotAuthenticated
	}

	requestID := r.nextRequestID()

	// Command packet
	// https://wiki.vg/RCON#2:_Command
	{
		buf := &bytes.Buffer{}

		// Length - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(10+len(command))); err != nil {
			return err
		}

		// Request ID - int32
		if err := binary.Write(buf, binary.LittleEndian, requestID); err != nil {
			return err
		}

		// Type - int32
		if err := binary.Write(buf, binary.LittleEndian, int32(2)); err != nil {
			return err
		}

		// Payload - null-terminated string
		if _, err := buf.Write(append([]byte(command), 0x00)); err != nil {
			return err
		}

		if err := buf.WriteByte(0x00); err != nil {
			return err
		}

		if _, err := io.Copy(r.conn, buf); err != nil {
			return err
		}
	}

	return nil
}

// Execute runs the command on the server and attempts to wait for a response.
func (r *Client) Execute(ctx context.Context, command string) (string, error) {
	if r.conn == nil {
		return "", ErrNotConnected
	}

	if !r.authSuccess {
		return "", ErrNotAuthenticated
	}

	if err := r.Run(command); err != nil {
		return "", err
	}

	for {
		select {
		case v := <-r.Messages:
			return v, nil
		case <-ctx.Done():
			return "", context.DeadlineExceeded
		}
	}
}

// Close closes the connection to the server.
func (r *Client) Close() error {
	r.authSuccess = false
	r.requestID = 0

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}

	r.conn = nil

	return nil
}

func (r *Client) readMessage() error {
	// Command response packet
	// https://wiki.vg/RCON#0:_Command_response
	{
		var (
			packetLength int32
			requestID    int32
		)

		// Length - int32
		{
			// TODO convert to binary.Read() for the rest of the package
			data := make([]byte, 4)

			n, err := r.conn.Read(data)

			if err != nil {
				return err
			}

			if n < 4 {
				return nil
			}

			packetLength = int32(binary.LittleEndian.Uint16(data))
		}

		// Request ID - int32
		{
			if err := binary.Read(r.conn, binary.LittleEndian, &requestID); err != nil {
				return err
			}
		}

		// Type - int32
		{
			var packetType int32

			if err := binary.Read(r.conn, binary.LittleEndian, &packetType); err != nil {
				return err
			}

			if packetType != 0 {
				return fmt.Errorf("rcon: received unexpected packet type (expected=0x00, received=0x%02X)", packetType)
			}
		}

		// Payload - null-terminated string
		{
			data := make([]byte, packetLength-8)

			n, err := r.conn.Read(data)

			if err != nil {
				return err
			}

			if n < 1 {
				return nil
			}

			r.Messages <- string(data)
		}
	}

	return nil
}

func (r *Client) nextRequestID() int32 {
	value := r.requestID

	r.requestID++

	return value
}

func parseOptions(opts ...options.RCON) options.RCON {
	if len(opts) < 1 {
		return defaultOptions
	}

	return opts[0]
}
