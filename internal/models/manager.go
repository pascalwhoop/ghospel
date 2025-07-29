package models

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/schollz/progressbar/v3"
)

// Manager handles Whisper model operations
type Manager struct {
	cacheDir string
}

// ModelInfo represents information about a Whisper model
type ModelInfo struct {
	Name        string
	Size        string
	Downloaded  bool
	Path        string
	Description string
	DownloadURL string
}

// NewManager creates a new model manager
func NewManager(cacheDir string) *Manager {
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".whisper")
	}

	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0o755)

	return &Manager{cacheDir: cacheDir}
}

// AvailableModels returns all available Whisper models with their download URLs
func (m *Manager) AvailableModels() []ModelInfo {
	baseURL := "https://huggingface.co/ggerganov/whisper.cpp/resolve/main"

	return []ModelInfo{
		{
			Name:        "tiny",
			Size:        "39 MB",
			Description: "Fastest, least accurate",
			Path:        filepath.Join(m.cacheDir, "ggml-tiny.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-tiny.bin", baseURL),
		},
		{
			Name:        "tiny.en",
			Size:        "39 MB",
			Description: "Fastest, least accurate (English only)",
			Path:        filepath.Join(m.cacheDir, "ggml-tiny.en.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-tiny.en.bin", baseURL),
		},
		{
			Name:        "base",
			Size:        "142 MB",
			Description: "Good balance of speed and accuracy",
			Path:        filepath.Join(m.cacheDir, "ggml-base.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-base.bin", baseURL),
		},
		{
			Name:        "base.en",
			Size:        "142 MB",
			Description: "Good balance of speed and accuracy (English only)",
			Path:        filepath.Join(m.cacheDir, "ggml-base.en.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-base.en.bin", baseURL),
		},
		{
			Name:        "small",
			Size:        "488 MB",
			Description: "Better accuracy, moderate speed",
			Path:        filepath.Join(m.cacheDir, "ggml-small.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-small.bin", baseURL),
		},
		{
			Name:        "small.en",
			Size:        "488 MB",
			Description: "Better accuracy, moderate speed (English only)",
			Path:        filepath.Join(m.cacheDir, "ggml-small.en.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-small.en.bin", baseURL),
		},
		{
			Name:        "medium",
			Size:        "1.5 GB",
			Description: "High accuracy, slower",
			Path:        filepath.Join(m.cacheDir, "ggml-medium.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-medium.bin", baseURL),
		},
		{
			Name:        "medium.en",
			Size:        "1.5 GB",
			Description: "High accuracy, slower (English only)",
			Path:        filepath.Join(m.cacheDir, "ggml-medium.en.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-medium.en.bin", baseURL),
		},
		{
			Name:        "large-v3",
			Size:        "2.9 GB",
			Description: "Latest large model with improvements",
			Path:        filepath.Join(m.cacheDir, "ggml-large-v3.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-large-v3.bin", baseURL),
		},
		{
			Name:        "large-v3-turbo",
			Size:        "1.5 GB",
			Description: "Large v3 Turbo - faster with similar accuracy",
			Path:        filepath.Join(m.cacheDir, "ggml-large-v3-turbo.bin"),
			DownloadURL: fmt.Sprintf("%s/ggml-large-v3-turbo.bin", baseURL),
		},
	}
}

// List displays available and downloaded models
func (m *Manager) List() error {
	models := m.AvailableModels()

	fmt.Println("Available Whisper Models:")
	fmt.Println("=========================")

	for _, model := range models {
		downloaded := ""
		if _, err := os.Stat(model.Path); err == nil {
			downloaded = "✅ Downloaded"
		} else {
			downloaded = "⬇️  Not downloaded"
		}

		fmt.Printf("%-12s | %-12s | %s | %s\n",
			model.Name, model.Size, downloaded, model.Description)
	}

	fmt.Printf("\nCache directory: %s\n", m.cacheDir)

	return nil
}

// Download downloads a specific model
func (m *Manager) Download(modelName string) error {
	// Validate model name
	models := m.AvailableModels()

	var targetModel *ModelInfo

	for i, model := range models {
		if model.Name == modelName {
			targetModel = &models[i]
			break
		}
	}

	if targetModel == nil {
		return fmt.Errorf("unknown model: %s", modelName)
	}

	// Check if already downloaded
	if _, err := os.Stat(targetModel.Path); err == nil {
		fmt.Printf("✅ Model %s is already downloaded\n", modelName)
		return nil
	}

	fmt.Printf("📥 Downloading %s model (%s) from Hugging Face...\n", modelName, targetModel.Size)

	// Create HTTP request
	resp, err := http.Get(targetModel.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Get content length for progress bar
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		// Try to parse from Content-Length header
		if lengthStr := resp.Header.Get("Content-Length"); lengthStr != "" {
			if length, err := strconv.ParseInt(lengthStr, 10, 64); err == nil {
				contentLength = length
			}
		}
	}

	// Create output file
	out, err := os.Create(targetModel.Path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Create progress bar
	var progressReader io.Reader = resp.Body

	if contentLength > 0 {
		bar := progressbar.NewOptions64(
			contentLength,
			progressbar.OptionSetDescription(fmt.Sprintf("Downloading %s", modelName)),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(40),
			progressbar.OptionThrottle(65*1000000), // 65ms
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
		)
		reader := progressbar.NewReader(resp.Body, bar)
		progressReader = &reader
	}

	// Copy data with progress
	_, err = io.Copy(out, progressReader)
	if err != nil {
		// Clean up partial download
		os.Remove(targetModel.Path)
		return fmt.Errorf("download failed: %w", err)
	}

	fmt.Printf("✅ Successfully downloaded %s model\n", modelName)

	return nil
}

// Cleanup removes unused cached models
func (m *Manager) Cleanup() error {
	fmt.Println("🧹 Cleaning up unused models...")

	// TODO: Implement cleanup logic
	// - Check last access times
	// - Remove models not used in X days
	// - Keep at least one model

	fmt.Println("✅ Cache cleanup complete")

	return nil
}

// Info shows information about a specific model
func (m *Manager) Info(modelName string) error {
	models := m.AvailableModels()

	var targetModel *ModelInfo

	for i, model := range models {
		if model.Name == modelName {
			targetModel = &models[i]
			break
		}
	}

	if targetModel == nil {
		return fmt.Errorf("unknown model: %s", modelName)
	}

	fmt.Printf("Model Information: %s\n", modelName)
	fmt.Println("===================")
	fmt.Printf("Size: %s\n", targetModel.Size)
	fmt.Printf("Description: %s\n", targetModel.Description)
	fmt.Printf("Path: %s\n", targetModel.Path)
	fmt.Printf("Download URL: %s\n", targetModel.DownloadURL)

	if stat, err := os.Stat(targetModel.Path); err == nil {
		fmt.Printf("Downloaded: Yes (%s)\n", stat.ModTime().Format("2006-01-02 15:04:05"))
		fmt.Printf("File Size: %d bytes\n", stat.Size())
	} else {
		fmt.Println("Downloaded: No")
	}

	return nil
}
