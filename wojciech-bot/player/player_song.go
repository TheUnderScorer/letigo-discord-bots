package player

import "time"

type Song struct {
	Url          string
	Name         string
	AuthorID     string
	Duration     time.Duration
	ThumbnailUrl string
}
