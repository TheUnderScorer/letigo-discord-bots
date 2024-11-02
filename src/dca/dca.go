package dca

import (
	"github.com/jonas747/dca/v2"
	"io"
)

func Convert(path string) (*dca.EncodeSession, error) {
	session, err := dca.EncodeFile(path, dca.StdEncodeOptions)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func ConvertReader(data io.Reader) (*dca.EncodeSession, error) {
	session, err := dca.EncodeMem(data, dca.StdEncodeOptions)
	if err != nil {
		return nil, err
	}

	return session, nil
}
