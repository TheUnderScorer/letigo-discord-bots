package util

import "math/rand"

func RandomInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func RandomBool() bool {
	return rand.Intn(2) == 1
}
