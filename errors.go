package mcutil

import "errors"

var (
	// ErrNotConnected means the client attempted to send data but there was no connection to the server
	ErrNotConnected = errors.New("rcon: not connected to the server")
	// ErrAlreadyLoggedIn means the RCON client was already logged in but a second login attempt was made
	ErrAlreadyLoggedIn = errors.New("rcon: already successfully logged in")
	// ErrInvalidPassword means the password used in the RCON login was incorrect
	ErrInvalidPassword = errors.New("rcon: incorrect password")
	// ErrNotAuthenticated means the client attempted to execute a command before a login was successful
	ErrNotAuthenticated = errors.New("rcon: not authenticated with the server")
)
