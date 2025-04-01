package arrayutil

import (
	"app/util"
	"math/rand"
)

func Includes[T comparable](arr []T, val T) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func Delete[T any](arr []T, i int) []T {
	return append(arr[:i], arr[i+1:]...)
}

func Last[T any](arr []T) (T, bool) {
	if len(arr) == 0 {
		return *new(T), false
	}

	return arr[len(arr)-1], true
}

func FindLast[T any](arr []T, predicate func(T) bool) (T, bool) {
	if len(arr) == 0 {
		return *new(T), false
	}

	reversedArr := make([]T, len(arr))
	copy(reversedArr, arr)

	return Find(reversedArr, predicate)
}

func Shuffle[T any](arr []T) []T {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}

func Filter[T any](arr []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range arr {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
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

	index := util.RandomInt(0, len(arr))

	return arr[index]
}

func Map[T any, U any](arr []T, mapper func(T) U) []U {
	result := make([]U, len(arr))
	for i, v := range arr {
		result[i] = mapper(v)
	}
	return result
}

// ReverseSlice reverses any slice
func ReverseSlice[T any](slice []T) []T {
	sliceCopy := make([]T, len(slice), len(slice))

	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		sliceCopy[i], sliceCopy[j] = slice[j], slice[i]
	}

	return sliceCopy
}
