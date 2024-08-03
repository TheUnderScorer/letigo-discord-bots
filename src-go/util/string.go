package util

import (
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
