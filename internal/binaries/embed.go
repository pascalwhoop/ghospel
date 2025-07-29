//go:build release

package binaries

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Embedded binaries for different platforms
// These will be populated by the build system

//go:embed all:whisper-cli-*
var embeddedFS embed.FS

// ExtractWhisperBinary extracts the appropriate whisper binary to a temporary location
func ExtractWhisperBinary() (string, error) {
	// Determine the correct binary for current platform
	filename := fmt.Sprintf("whisper-cli-%s-%s", runtime.GOOS, runtime.GOARCH)
	
	// Check if the binary exists in the embedded filesystem
	binaryData, err := embeddedFS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("binary not embedded for platform %s-%s: %w", runtime.GOOS, runtime.GOARCH, err)
	}
	
	// Create temporary directory for the binary
	tmpDir, err := os.MkdirTemp("", "ghospel-whisper-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Write binary to temp file
	binaryPath := filepath.Join(tmpDir, filename)
	file, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to create binary file: %w", err)
	}
	defer file.Close()
	
	_, err = file.Write(binaryData)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to write binary: %w", err)
	}
	
	return binaryPath, nil
}

// IsEmbeddedBinaryAvailable checks if a binary is available for the current platform
func IsEmbeddedBinaryAvailable() bool {
	filename := fmt.Sprintf("whisper-cli-%s-%s", runtime.GOOS, runtime.GOARCH)
	_, err := embeddedFS.ReadFile(filename)
	return err == nil
}