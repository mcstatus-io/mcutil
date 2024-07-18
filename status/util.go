package status

import (
	"bytes"
	"fmt"
	"io"

	"github.com/mcstatus-io/mcutil/v4/proto"
)

func writePacket(w io.Writer, data *bytes.Buffer) error {
	if err := proto.WriteVarInt(int32(data.Len()), w); err != nil {
		return err
	}

	_, err := io.Copy(w, data)

	return err
}

func parsePlayerID(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []interface{}:
		{
			if len(v) != 4 {
				return "", false
			}

			var a, b uint64

			for i, val := range v {
				parsed, ok := val.(float64)

				if !ok {
					return "", false
				}

				if i < 2 {
					a |= uint64(parsed) << ((i % 2) * 8)
				} else {
					b |= uint64(parsed) << ((i % 2) * 8)
				}
			}

			return fmt.Sprintf("%016x%016x", a, b), true
		}
	default:
		return "", false
	}
}

func pointerOf[T any](v T) *T {
	return &v
}
