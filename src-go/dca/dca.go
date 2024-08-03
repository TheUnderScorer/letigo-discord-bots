package dca

import (
	"github.com/jonas747/dca/v2"
)

func Convert(path string) (*dca.EncodeSession, error) {
	session, err := dca.EncodeFile(path, dca.StdEncodeOptions)
	if err != nil {
		return nil, err
	}

	return session, nil
}
