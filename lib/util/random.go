package util

import (
	"math/rand"
	"time"
)

func RandomInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func RandomBool() bool {
	return rand.Intn(2) == 1
}

func RandomDuration(min, max time.Duration) time.Duration {
	maxSeconds := max.Seconds()

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		n := random.Intn(int(maxSeconds))

		duration := time.Duration(n) * time.Second
		if duration >= min && duration <= max {
			return duration
		}
	}
}
