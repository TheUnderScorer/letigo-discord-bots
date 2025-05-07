package random

import (
	"crypto/rand"
	"encoding/binary"
	"math"
)

// ChanceOfTrue returns true with the given probability (0.0 to 1.0)
// and false otherwise, using cryptographically secure random numbers
func ChanceOfTrue(probability float64) bool {
	// Ensure probability is within valid range
	if probability <= 0.0 {
		return false
	}
	if probability >= 1.0 {
		return true
	}

	// Generate a random 64-bit value
	var randomBytes [8]byte
	_, err := rand.Read(randomBytes[:])
	if err != nil {
		// In case of failure, default to false
		return false
	}

	// Convert bytes to a float64 between 0 and 1
	randomInt := binary.LittleEndian.Uint64(randomBytes[:])
	// Convert to a float between 0 and 1
	randomFloat := float64(randomInt) / float64(math.MaxUint64)

	// Compare with the given probability
	return randomFloat < probability
}
