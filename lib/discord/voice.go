package discord

import (
	"github.com/jonas747/dca/v2"
)

type Voice struct {
	reader dca.OpusReader

	framesSentCh chan<- int
	framesSent   int
}

func NewVoice(reader dca.OpusReader) *Voice {
	return &Voice{
		reader:     reader,
		framesSent: 0,
	}
}

func NewFrameTrackedVoice(reader dca.OpusReader, framesSentCh chan<- int) *Voice {
	return &Voice{
		reader:       reader,
		framesSent:   0,
		framesSentCh: framesSentCh,
	}
}

func (v *Voice) SetFramesSentCh(framesSentCh chan<- int) {
	v.framesSentCh = framesSentCh
}

func (v *Voice) Reader() dca.OpusReader {
	return v.reader
}
