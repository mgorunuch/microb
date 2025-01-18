package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Create a map to track unique lines
	seen := make(map[string]bool)

	// Create scanner to read stdin line by line
	scanner := bufio.NewScanner(os.Stdin)

	// Scan each line
	for scanner.Scan() {
		line := scanner.Text()

		// Only output if we haven't seen this line before
		if !seen[line] {
			fmt.Println(line)
			seen[line] = true
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}
}
