package random_test

import (
	"github.com/stretchr/testify/assert"
	"lib/random"
	"testing"
)

func TestChanceOfTrue(t *testing.T) {
	t.Run("with 100% chance", func(t *testing.T) {
		result := random.ChanceOfTrue(1)

		assert.True(t, result)
	})

	t.Run("with 0% chance", func(t *testing.T) {
		result := random.ChanceOfTrue(0)

		assert.False(t, result)
	})

	t.Run("Statistical distribution test", func(t *testing.T) {
		// Number of iterations
		const iterations = 10000
		const probability = 0.7
		const acceptableDeviationPercent = 5.0 // Allow 5% deviation

		// Count true results
		trueCount := 0
		for i := 0; i < iterations; i++ {
			if random.ChanceOfTrue(probability) {
				trueCount++
			}
		}

		// Calculate actual percentage
		actualProb := float64(trueCount) / float64(iterations)

		// Define acceptable range
		lowerBound := probability * (1 - acceptableDeviationPercent/100)
		upperBound := probability * (1 + acceptableDeviationPercent/100)

		if actualProb < lowerBound || actualProb > upperBound {
			t.Errorf("Expected probability around %.2f, but got %.2f (outside acceptable range of %.2f to %.2f)",
				probability, actualProb, lowerBound, upperBound)
		}
	})

}
