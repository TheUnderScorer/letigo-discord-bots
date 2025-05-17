package progress

import "github.com/schollz/progressbar/v3"

type Bar struct {
	Total int64
	Value int64

	progress *progressbar.ProgressBar
}

func NewBar(total int64) *Bar {
	progress := progressbar.NewOptions64(total,
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(""),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetMaxDetailRow(1),
		progressbar.OptionSetWidth(20),
		progressbar.OptionUseANSICodes(true),
	)

	return &Bar{
		Total: total,
		Value: 0,

		progress: progress,
	}
}

func (b *Bar) SetValue(value int64) error {
	b.Value = value

	return b.progress.Set64(value)
}

func (b *Bar) String() string {
	return b.progress.String()
}
