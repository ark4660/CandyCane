package components

import (
	"crypto/rand"
    "encoding/hex"
)

func GenerateSessionId() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil // 64-character hex string
}
