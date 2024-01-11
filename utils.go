package bluerpc

// findIndex finds the index of a string in a slice. It returns -1 if the string is not found.
func findIndex(slice []string, val string) int {
	for i, item := range slice {
		if item == val {
			return i
		}
	}
	return -1
}
