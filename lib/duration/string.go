package duration

import (
	"fmt"
	"time"
)

// ToMinSec converts a time.Duration into a string formatted as "MM:SS", representing minutes and seconds.
func ToMinSec(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	m := totalSeconds / 60
	s := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d", m, s)
}
