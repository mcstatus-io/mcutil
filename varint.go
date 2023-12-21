package mcutil

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	// ErrVarIntTooBig means the varint received from the server is too big
	ErrVarIntTooBig = errors.New("varint: varint is too big")
)

func readVarInt(r io.Reader) (int32, error) {
	var value int32 = 0
	var position int = 0
	var currentByte byte

	for {
		if err := binary.Read(r, binary.BigEndian, &currentByte); err != nil {
			return 0, err
		}

		value |= int32(currentByte&0x7F) << position

		if currentByte&0x80 == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, ErrVarIntTooBig
		}
	}

	return value, nil
}

func writeVarInt(val int32, w io.Writer) error {
	for {
		if (val & 0x80) == 0 {
			_, err := w.Write([]byte{byte(val)})

			return err
		}

		if _, err := w.Write([]byte{byte((val & 0x7F) | 0x80)}); err != nil {
			return err
		}

		val = int32(uint32(val) >> 7)
	}
}
