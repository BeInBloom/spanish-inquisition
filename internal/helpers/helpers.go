package helpers

import "strings"

func GetParams(path string) []string {
	return filter(strings.Split(path, "/"), func(s string) bool {
		return s != ""
	})
}

func filter[T any](slice []T, f func(T) bool) []T {
	result := make([]T, 0)

	for _, item := range slice {
		if f(item) {
			result = append(result, item)
		}
	}

	return result
}
