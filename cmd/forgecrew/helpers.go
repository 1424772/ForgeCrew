package main

import (
	"fmt"
	"os"
)

// writeIfMissing writes content to path only if the file doesn't exist or force is true.
func writeIfMissing(path, content string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  skipping %s (already exists, use --force to overwrite)\n", path)
			return nil
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
