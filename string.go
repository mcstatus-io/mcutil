package mcstatus

import (
	"io"
)

func readString(r io.Reader) ([]byte, error) {
	length, _, err := readVarInt(r)

	if err != nil {
		return nil, err
	}

	data := make([]byte, length)

	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return data, nil
}

func writeString(val string, w io.Writer) error {
	if _, err := writeVarInt(int32(len(val)), w); err != nil {
		return err
	}

	_, err := w.Write([]byte(val))

	return err
}
