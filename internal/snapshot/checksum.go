package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// CalculateChecksum generates a SHA256 checksum for any data
func CalculateChecksum(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyChecksum verifies if data matches the expected checksum
func VerifyChecksum(data interface{}, expected string) (bool, error) {
	actual, err := CalculateChecksum(data)
	if err != nil {
		return false, err
	}
	return actual == expected, nil
}
