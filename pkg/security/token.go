package security

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
