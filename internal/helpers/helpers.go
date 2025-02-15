package helpers

import (
	"math/rand/v2"
	"strings"
)

func GetParams(path string) []string {
	return Filter(strings.Split(path, "/"), func(s string) bool {
		return s != ""
	})
}

func Filter[T any](slice []T, f func(T) bool) []T {
	result := make([]T, 0)

	for _, item := range slice {
		if f(item) {
			result = append(result, item)
		}
	}

	return result
}

func GetRandomFloat(min, max float64) *float64 {
	res := min + (max-min)*rand.Float64()
	return &res
}
