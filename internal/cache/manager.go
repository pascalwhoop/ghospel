package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles cache operations
type Manager struct {
	cacheDir string
}

// NewManager creates a new cache manager
func NewManager(cacheDir string) *Manager {
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".whisper")
	}

	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0o755)

	return &Manager{cacheDir: cacheDir}
}

// Info displays cache statistics
func (m *Manager) Info() error {
	fmt.Println("Cache Information:")
	fmt.Println("==================")

	// Calculate cache size
	var totalSize int64

	var fileCount int

	err := filepath.Walk(m.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to calculate cache size: %w", err)
	}

	fmt.Printf("Location: %s\n", m.cacheDir)
	fmt.Printf("Total Size: %s\n", formatBytes(totalSize))
	fmt.Printf("File Count: %d\n", fileCount)

	// Check if cache directory exists
	if _, err := os.Stat(m.cacheDir); os.IsNotExist(err) {
		fmt.Println("Status: Cache directory does not exist")
	} else {
		fmt.Println("Status: Active")
	}

	return nil
}

// Clean removes old cached files
func (m *Manager) Clean(olderThan string) error {
	fmt.Printf("🧹 Cleaning cache files older than %s...\n", olderThan)

	// Parse duration
	duration, err := parseDuration(olderThan)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	cutoff := time.Now().Add(-duration)

	var removedCount int

	var removedSize int64

	err = filepath.Walk(m.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and recently modified files
		if info.IsDir() || info.ModTime().After(cutoff) {
			return nil
		}

		// Don't remove model files during clean (only during clear)
		if filepath.Dir(path) == filepath.Join(m.cacheDir, "models") {
			return nil
		}

		removedSize += info.Size()
		removedCount++

		return os.Remove(path)
	})
	if err != nil {
		return fmt.Errorf("failed to clean cache: %w", err)
	}

	fmt.Printf("✅ Removed %d files (%s freed)\n", removedCount, formatBytes(removedSize))

	return nil
}

// Clear removes all cached files
func (m *Manager) Clear(force bool) error {
	if !force {
		fmt.Print("⚠️  This will remove all cached files including models. Continue? (y/N): ")

		var response string

		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	fmt.Println("🗑️  Clearing entire cache...")

	// Remove entire cache directory
	if err := os.RemoveAll(m.cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	// Recreate empty cache directory
	if err := os.MkdirAll(m.cacheDir, 0o755); err != nil {
		return fmt.Errorf("failed to recreate cache directory: %w", err)
	}

	fmt.Println("✅ Cache cleared successfully")

	return nil
}

// ShowPath displays the cache directory path
func (m *Manager) ShowPath() error {
	fmt.Println(m.cacheDir)
	return nil
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseDuration parses duration strings like "30d", "7d", "24h"
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	unit := s[len(s)-1]
	value := s[:len(s)-1]

	switch unit {
	case 'd':
		// Parse as days
		if n := parseInt(value); n > 0 {
			return time.Duration(n) * 24 * time.Hour, nil
		}
	case 'h':
		// Parse as hours
		if n := parseInt(value); n > 0 {
			return time.Duration(n) * time.Hour, nil
		}
	}

	// Fallback to standard time.ParseDuration
	return time.ParseDuration(s)
}

// parseInt is a simple integer parser
func parseInt(s string) int {
	n := 0

	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0
		}
	}

	return n
}
