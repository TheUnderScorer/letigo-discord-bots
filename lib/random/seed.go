package random

import (
	"math/rand"
	"time"
)

func Seed() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
