package mcutil

import "errors"

var (
	// ErrVarIntTooBig means the varint received from the server is too big
	ErrVarIntTooBig = errors.New("varint: too big, overflows")
)
