package main

import (
	"fmt"
	"io"
	"os"

	"github.com/1424772/ForgeCrew/internal/i18n"
)

// writeIfMissing writes content to path only if the file doesn't exist or force is true.
func writeIfMissing(w io.Writer, path, content string, force bool, loc i18n.Locale) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintln(w, i18n.T("init.skipping", loc)+path)
			return nil
		}
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
