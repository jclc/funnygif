package funnygif

import (
	"os"
	"path/filepath"
)

func getFontPaths() []string {
	return []string{"/usr/share/fonts", filepath.Join(os.Getenv("HOME"), ".fonts")}
}
