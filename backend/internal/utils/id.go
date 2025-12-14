package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
)

// GenerateID генерирует случайный ID
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateUUID генерирует UUID v4
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	// Устанавливаем версию (4) и вариант UUID
	b[6] = (b[6] & 0x0f) | 0x40 // версия 4
	b[8] = (b[8] & 0x3f) | 0x80 // вариант
	return hex.EncodeToString(b[0:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:16])
}

// UUIDToInt64 конвертирует UUID в int64 для использования в math/rand
func UUIDToInt64(uuid string) int64 {
	h := fnv.New64a()
	h.Write([]byte(uuid))
	return int64(h.Sum64())
}

// Int64ToUUID конвертирует int64 обратно в UUID-подобную строку (для временной совместимости с protobuf)
func Int64ToUUID(val int64) string {
	// Временная функция для обратной конвертации
	// После перегенерации protobuf эта функция не понадобится
	return fmt.Sprintf("%016x-%04x-%04x-%04x-%012x", val, val>>48, val>>32, val>>16, val)
}

