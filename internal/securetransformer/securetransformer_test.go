package securetransformer

import (
	"testing"
)

// TestBasicEncoding ensures Encode produces a non-empty string for a valid int64 input
func TestBasicEncoding(t *testing.T) {
	transformer := NewSecureTransformer("test_seed_123")
	value := int64(123456789)
	encoded := transformer.Encode(value)

	if encoded == "" {
		t.Errorf("Encode() returned an empty string for input %d", value)
	}
}

// TestDifferentSeeds ensures that different seeds produce different encoded outputs
func TestDifferentSeeds(t *testing.T) {
	transformer1 := NewSecureTransformer("seed_one")
	transformer2 := NewSecureTransformer("seed_two")

	value := int64(123456789)
	encoded1 := transformer1.Encode(value)
	encoded2 := transformer2.Encode(value)

	if encoded1 == encoded2 {
		t.Errorf("Different seeds should produce different outputs")
	}
}

// TestSameSeedSameOutput ensures that the same seed and same input always produce the same result
func TestSameSeedSameOutput(t *testing.T) {
	seed := "consistent_seed"
	transformer := NewSecureTransformer(seed)

	value := int64(123456789)
	encoded1 := transformer.Encode(value)
	encoded2 := transformer.Encode(value)

	if encoded1 != encoded2 {
		t.Errorf("Encode() is not deterministic. Expected %s, got %s", encoded1, encoded2)
	}
}

// TestDifferentInputs ensures that different int64 values produce different encoded outputs
func TestDifferentInputs(t *testing.T) {
	transformer := NewSecureTransformer("test_seed_123")

	encoded1 := transformer.Encode(int64(123456789))
	encoded2 := transformer.Encode(int64(987654321))

	if encoded1 == encoded2 {
		t.Errorf("Different inputs should produce different outputs")
	}
}

// TestEmptySeed ensures that creating a SecureTransformer with an empty seed returns an error
func TestEmptySeed(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for empty seed but got none")
		}
	}()

	NewSecureTransformer("") // Should panic due to empty seed
}
