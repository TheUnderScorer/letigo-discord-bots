package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	dca2 "github.com/jonas747/dca/v2"
	"io"
)

type NonDcaSpeaker struct {
	voice io.Reader
}

func NewNonDcaSpeaker(voice io.Reader) *NonDcaSpeaker {
	return &NonDcaSpeaker{voice: voice}
}

func (b NonDcaSpeaker) Speak(ctx context.Context, vc *discordgo.VoiceConnection) error {
	stream, err := dca2.EncodeMem(b.voice, dca2.StdEncodeOptions)
	if err != nil {
		return err
	}
	defer stream.Cleanup()

	err = vc.Speaking(true)
	defer func() {
		_ = vc.Speaking(false)
	}()
	if err != nil {
		return err
	}

	for {
		frame, err := stream.OpusFrame()
		if err != nil {
			break
		}

		select {
		case vc.OpusSend <- frame:
			continue
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
