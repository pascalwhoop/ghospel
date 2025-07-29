//go:build !release

package binaries

import (
	"fmt"
)

// ExtractWhisperBinary returns empty in development mode (binaries not embedded)
func ExtractWhisperBinary() (string, error) {
	return "", fmt.Errorf("embedded binaries not available in development mode")
}

// IsEmbeddedBinaryAvailable always returns false in development mode
func IsEmbeddedBinaryAvailable() bool {
	return false
}
