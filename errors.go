package mcstatus

import "errors"

var (
	// ErrUnexpectedResponse means the server sent an unexpected response to the client
	ErrUnexpectedResponse = errors.New("received an unexpected response from the server")
	// ErrVarIntTooBig means the server sent a varint which was beyond the protocol size of a varint
	ErrVarIntTooBig = errors.New("size of VarInt exceeds maximum data size")
	// ErrNotConnected means the client attempted to send data but there was no connection to the server
	ErrNotConnected = errors.New("client attempted to send data but connection is non-existent")
	// ErrAlreadyLoggedIn means the RCON client was already logged in after a second login attempt was made
	ErrAlreadyLoggedIn = errors.New("RCON client is already logged in after a second login attempt was made")
	// ErrInvalidPassword means the password used in the RCON loggin was incorrect
	ErrInvalidPassword = errors.New("incorrect RCON password")
	// ErrNotLoggedIn means the client attempted to execute a command before a login was successful
	ErrNotLoggedIn = errors.New("RCON client attempted to send message before successful login")
	// ErrDecodeUTF16OddLength means a UTF-16 was attempted to be decoded from a byte array that was an odd length
	ErrDecodeUTF16OddLength = errors.New("attempted to decode UTF-16 byte array with an odd length")
)
