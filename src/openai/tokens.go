package openai

import "github.com/pkoukk/tiktoken-go"

// CountTokens computes the number of tokens in the given contents using the specified encodingName.
// Returns the token count as int32 and an error if the encoding fails.
func CountTokens(contents string, encodingName string) (int32, error) {
	tke, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return 0, err
	}

	token := tke.Encode(contents, nil, nil)
	return int32(len(token)), nil
}
