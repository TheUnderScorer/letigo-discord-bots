package player

import yt "github.com/kkdai/youtube/v2"

type Song struct {
	Url       string
	StreamUrl string
	VideoID   string
	Name      string
	AuthorID  string
	Format    *yt.Format
}
