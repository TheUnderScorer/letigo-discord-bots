package ytdlp

import (
	"os/exec"
)

const cli = "yt-dlp"

// getCommand constructs an exec.Cmd to execute the yt-dlp command-line tool with the provided URL and additional arguments.
func getCommand(url string, additionalArgs ...string) *exec.Cmd {
	cookiesArgs := getCookiesArgs()

	var args []string

	if len(cookiesArgs) > 0 {
		args = append(args, cookiesArgs...)
	}

	args = append(args, additionalArgs...)
	args = append(args, url)

	cmd := exec.Command(cli, args...)

	return cmd
}
