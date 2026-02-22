package api

import (
	"strings"
)

// isNotFoundError checks if an error indicates a resource was not found
func isNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "no rows in result set") ||
		strings.Contains(errMsg, "no such row") ||
		strings.Contains(errMsg, "not found")
}
