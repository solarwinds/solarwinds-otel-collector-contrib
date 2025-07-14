package internal

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
)

func HashObject(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to encode: %w", err)
	}

	// Create hash from the marshaled bytes
	h := fnv.New64a()
	_, err = h.Write(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to write bytes to hash: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}
