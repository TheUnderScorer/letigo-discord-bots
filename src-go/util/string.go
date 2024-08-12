package util

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func ApplyTokens(input string, tokens map[string]string) string {
	for k, v := range tokens {
		tokenInString := fmt.Sprintf(`{{%s}}`, k)

		input = strings.ReplaceAll(input, tokenInString, v)
	}
	return input
}

// Hash hashes a string to a hex string using sha256
func Hash(input string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}
