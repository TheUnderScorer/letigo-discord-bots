package voice

import (
	"context"
	"github.com/bwmarrin/discordgo"
	dca2 "github.com/jonas747/dca/v2"
	"io"
)

type DcaSpeaker struct {
	voice io.Reader
}

func NewDcaSpeaker(voice io.Reader) *DcaSpeaker {
	return &DcaSpeaker{voice: voice}
}

func (d *DcaSpeaker) Speak(ctx context.Context, vc *discordgo.VoiceConnection) error {
	decoder := dca2.NewDecoder(d.voice)

	err := vc.Speaking(true)
	defer func() {
		_ = vc.Speaking(false)
	}()
	if err != nil {
		return err
	}

	for {
		frame, err := decoder.OpusFrame()

		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		select {
		case vc.OpusSend <- frame:
			continue
		case <-ctx.Done():
			return nil
		}
	}
}
