package generators

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
)

// KeyGenerator defines the interface for generating keys
type KeyGenerator interface {
	Generate(index int) string
}

// IntegerKeyGenerator generates integer keys
type IntegerKeyGenerator struct{}

// Generate creates a new integer key
func (g *IntegerKeyGenerator) Generate(index int) string {
	return strconv.Itoa(index)
}

// StringKeyGenerator generates string keys with specified length
type StringKeyGenerator struct {
	Length int
}

// Generate creates a new string key
func (g *StringKeyGenerator) Generate(index int) string {
	return RandomString(g.Length)
}

// UUIDKeyGenerator generates UUID keys
type UUIDKeyGenerator struct{}

// Generate creates a new UUID key
func (g *UUIDKeyGenerator) Generate(index int) string {
	return uuid.New().String()
}

// NewKeyGenerator creates a new key generator based on the key type
func NewKeyGenerator(keyType string) (KeyGenerator, error) {
	switch keyType {
	case "integer":
		return &IntegerKeyGenerator{}, nil
	case "string26":
		return &StringKeyGenerator{Length: 26}, nil
	case "string90":
		return &StringKeyGenerator{Length: 90}, nil
	case "string250":
		return &StringKeyGenerator{Length: 250}, nil
	case "string506":
		return &StringKeyGenerator{Length: 506}, nil
	case "uuid":
		return &UUIDKeyGenerator{}, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// GenerateKeys generates a slice of keys
func GenerateKeys(keyType string, count int, random bool) ([]string, error) {
	generator, err := NewKeyGenerator(keyType)
	if err != nil {
		return nil, err
	}

	keys := make([]string, count)
	indices := make([]int, count)
	
	// Create sequential or random indices
	for i := 0; i < count; i++ {
		indices[i] = i
	}
	
	// Randomize indices if requested
	if random {
		rand.Shuffle(count, func(i, j int) {
			indices[i], indices[j] = indices[j], indices[i]
		})
	}
	
	// Generate keys
	for i := 0; i < count; i++ {
		keys[i] = generator.Generate(indices[i])
	}
	
	return keys, nil
} 