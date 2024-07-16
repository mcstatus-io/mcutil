package query

import "io"

func readNTString(r io.Reader) (string, error) {
	result := make([]byte, 0)

	for {
		data := make([]byte, 1)

		if _, err := r.Read(data); err != nil {
			return "", err
		}

		if data[0] == 0x00 {
			break
		}

		result = append(result, data...)
	}

	return string(result), nil
}
