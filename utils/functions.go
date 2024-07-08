package utils

import (
	"strings"
)

func ContainsInt(list []int, num int) bool {
	for _, v := range list {
		if v == num {
			return true
		}
	}
	return false
}

// Helper function to check if a string is in a slice
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, item) { // case-insensitive comparison
			return true
		}
	}
	return false
}
