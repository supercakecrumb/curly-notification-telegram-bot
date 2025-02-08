package securetransformer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// SecureTransformer securely transforms an int64 into a non-reversible string
type SecureTransformer struct {
	seed []byte
}

// NewSecureTransformer initializes the transformer with a given seed
func NewSecureTransformer(seed string) *SecureTransformer {
	if len(seed) == 0 {
		panic("seed is empty")
	}
	return &SecureTransformer{seed: []byte(seed)}
}

// Encode securely transforms an int64 into a string
func (s *SecureTransformer) Encode(value int64) string {
	data := fmt.Sprintf("%d", value)

	// Generate HMAC-SHA256 using seed as key
	h := hmac.New(sha256.New, s.seed)
	h.Write([]byte(data))
	hash := h.Sum(nil)

	// Convert to Base64 for readability
	return base64.URLEncoding.EncodeToString(hash)
}

func main() {
	// Example usage
	seed := "my_super_secret_seed"
	transformer := NewSecureTransformer(seed)

	value := int64(123456789)
	encodedString := transformer.Encode(value)

	fmt.Println("Encoded:", encodedString)
}
