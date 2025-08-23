package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token), nil
}
