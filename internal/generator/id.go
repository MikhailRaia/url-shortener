package generator

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateID(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	id := base64.RawURLEncoding.EncodeToString(b)
	if len(id) > length {
		id = id[:length]
	}

	return id, nil
}
