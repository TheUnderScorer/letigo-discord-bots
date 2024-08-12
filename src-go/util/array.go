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

func Shuffle[T any](arr []T) []T {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
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
	if len(arr) == 1 {
		return arr[0]
	}

	index := RandomInt(0, len(arr))

	return arr[index]
}

func Map[T any, U any](arr []T, mapper func(T) U) []U {
	result := make([]U, len(arr))
	for i, v := range arr {
		result[i] = mapper(v)
	}
	return result
}
