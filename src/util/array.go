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

func RandomElements[T any](arr []T, count int) []T {
	if count >= len(arr) {
		return arr
	}

	return Shuffle(arr)[:count]
}

// IsValidArray checks if the provided array is non-nil and has a length greater than zero.
// It returns true if the array is valid, otherwise false.
func IsValidArray[T any](arr []T) bool {
	//nolint:all
	return arr != nil && len(arr) > 0
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

// ReverseSlice reverses any slice in-place
func ReverseSlice[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
