package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID генерирует случайный ID
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

