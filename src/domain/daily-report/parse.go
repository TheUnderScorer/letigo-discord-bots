package dailyreport

import (
	"fmt"
	"regexp"
	"strings"
	"util"
)

type DailyReport struct {
	Skipped bool
	Day     int
	Song    *Song
	Time    *TimeSpent
}

type Song struct {
	Url  string
	Name string
}

type TimeSpent struct {
	Seconds int
}

const freeDayToken = "wolne"
const songToken = "song dnia:"

func IsDailyReport(msg string) bool {
	split := strings.Split(strings.ToLower(msg), "\n")

	return util.Includes(split, "[dzień")
}

func Parse(msg string) {
	strings.Split(strings.ToLower(msg), "\n")
}

func ParseSong(msgSplit []string) *Song {
	line, isMatch := util.Find(msgSplit, func(s string) bool {
		return strings.Contains(s, songToken)
	})

	if !isMatch {
		return nil
	}

	re := regexp.MustCompile(fmt.Sprintf("/%s", songToken))

	songName := strings.ToLower(line)
	songName = re.ReplaceAllString(songName, "")

	if songName == "" {
		return nil
	}

	return &Song{}
}
