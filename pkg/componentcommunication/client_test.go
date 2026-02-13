package componentcommunication

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
)

func UUIDStringToBase64(s string) (string, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(id[:]), nil
}

func TestTransform(t *testing.T) {
	base64, err := UUIDStringToBase64("019c47bf-fea2-72fa-81b0-a2e9993ac601")
	if err != nil {
		t.Fatalf("Failed to transform UUID to Base64: %v", err)
	}
	t.Logf("Base64: %s", base64)
}
