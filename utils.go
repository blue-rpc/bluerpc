package bluerpc

import (
	"strings"
)

// findIndex finds the index of a string in a slice. It returns -1 if the string is not found.
func findIndex(slice []string, val string) int {
	for i, item := range slice {
		if item == val {
			return i
		}
	}
	return -1
}

// splitStringOnSlash splits the string at each slash and returns an array of substrings.
func splitStringOnSlash(s string) ([]string, error) {

	var result []string

	// Check if the string contains "/:" and find its position
	pos := strings.Index(s, "/:")
	if pos != -1 {
		// Split the string until the position of "/:"
		parts := strings.Split(s[:pos], "/")

		// Add non-empty parts to the result
		for _, part := range parts {
			if part != "" {
				result = append(result, "/"+part)
			}
		}

		// Add the remaining part of the string starting from "/:" as the last element
		result = append(result, s[pos:])
	} else {
		// Split the string at each slash if "/:" is not found
		parts := strings.Split(s, "/")

		// Add non-empty parts to the result
		for _, part := range parts {
			if part != "" {
				result = append(result, "/"+part)
			}
		}
	}

	return result, nil
}
