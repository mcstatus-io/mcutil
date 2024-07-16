package proto

import (
	"io"
)

func ReadString(r io.Reader) ([]byte, error) {
	length, err := ReadVarInt(r)

	if err != nil {
		return nil, err
	}

	data := make([]byte, length)

	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return data, nil
}

func WriteString(val string, w io.Writer) error {
	if err := WriteVarInt(int32(len(val)), w); err != nil {
		return err
	}

	_, err := w.Write([]byte(val))

	return err
}
