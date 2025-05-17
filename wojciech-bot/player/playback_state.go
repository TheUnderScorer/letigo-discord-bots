package player

import (
	"lib/duration"
	"lib/progress"
	"regexp"
	"strings"
	"time"
)

var percentageRegex = regexp.MustCompile(`\d+\.?\d*%`)

type playbackState struct {
	song              *Song
	isPlaying         func() bool
	remainingDuration time.Duration
	progressPlayed    int64

	progress *progress.Bar
}

func (p *playbackState) updateDuration(duration time.Duration) {
	p.remainingDuration = duration
}

func (p *playbackState) updateProgressPlayed(progress int64) error {
	return p.progress.SetValue(progress)
}

func (p *playbackState) Playing() bool {
	return p.isPlaying()
}

func (p *playbackState) String() string {
	progressBar := percentageRegex.ReplaceAllString(p.progress.String(), "")
	progressBar = strings.TrimSpace(progressBar)

	return duration.ToMinSec(p.remainingDuration) + " " + progressBar
}
