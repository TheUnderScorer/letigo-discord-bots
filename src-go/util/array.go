package util

import "math/rand"

func Includes[T comparable](arr []T, val T) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func Find[T any](arr []T, predicate func(T) bool) (T, bool) {
	for _, v := range arr {
		if predicate(v) {
			return v, true
		}
	}
	return *new(T), false
}

func RandomInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func RandomElement[T any](arr []T) T {
	index := RandomInt(0, len(arr)-1)

	return arr[index]
}
