package base64util

import "encoding/base64"

func EncodeBytes(data []byte) []byte {
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(b64, data)
	return b64
}
