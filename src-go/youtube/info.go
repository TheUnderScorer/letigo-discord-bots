package youtube

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type VideoInfo struct {
	Title string `json:"title"`
}

func GetVideoInfo(url string) (*VideoInfo, error) {
	// Get video metadata using yt-dlp
	cmd := exec.Command("yt-dlp", "-j", "--no-playlist", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// Parse JSON to extract the title
	var videoInfo VideoInfo
	err = json.Unmarshal(out.Bytes(), &videoInfo)
	if err != nil {
		return nil, err
	}

	return &videoInfo, nil
}
