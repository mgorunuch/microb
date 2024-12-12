package core

import "fmt"

// Helper function to parse int64 from string
func ParseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
