package internal

import (
	"encoding/base64"
	"os"
	"path"
)

func loadPem(value string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return data, nil
	}
	return os.ReadFile(path.Clean(value))
}
