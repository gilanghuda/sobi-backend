package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

func GenerateOTP(digits int) (string, error) {
	max := uint64(1)
	for i := 0; i < digits; i++ {
		max *= 10
	}

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	num := binary.LittleEndian.Uint64(b) % max
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, num), nil
}
