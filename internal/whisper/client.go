package whisper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pascalwhoop/ghospel/internal/binaries"
)

// Client provides a simple interface to whisper.cpp
type Client struct {
	whisperBinaryPath string
	modelsDir         string
}

// NewClient creates a new whisper client
func NewClient(whisperBinaryPath, modelsDir string) *Client {
	if whisperBinaryPath == "" {
		whisperBinaryPath = findWhisperBinary()
	}

	return &Client{
		whisperBinaryPath: whisperBinaryPath,
		modelsDir:         modelsDir,
	}
}

// findWhisperBinary attempts to locate the whisper binary in order of preference:
// 1. Embedded binary (release builds)
// 2. Development build location
// 3. System PATH
func findWhisperBinary() string {
	// First, try embedded binary (release builds)
	if binaries.IsEmbeddedBinaryAvailable() {
		if path, err := binaries.ExtractWhisperBinary(); err == nil {
			return path
		}
	}

	// Second, try development build location
	devPath := "./whisper_cpp_source/build/bin/whisper-cli"
	if _, err := os.Stat(devPath); err == nil {
		return devPath
	}

	// Third, try system PATH
	if path, err := exec.LookPath("whisper-cli"); err == nil {
		return path
	}

	// Fallback to development path (will fail gracefully if not found)
	return devPath
}

// Transcribe transcribes an audio file using the specified model
func (c *Client) Transcribe(audioPath, modelName string) (string, error) {
	// Construct model path
	modelPath := filepath.Join(c.modelsDir, fmt.Sprintf("ggml-%s.bin", modelName))

	// Build whisper command with Metal GPU acceleration (default enabled)
	cmd := exec.Command(c.whisperBinaryPath,
		"-m", modelPath, // Model path
		"-f", audioPath, // Audio file path
		"--output-txt",                         // Output as text
		"--output-file", "/tmp/ghospel_output", // Output file prefix
		"--language", "en", // Language (can be made configurable)
		"--threads", "4", // Number of threads
		"--flash-attn", // Enable flash attention for better performance
		// Note: --no-gpu is NOT used, so GPU/Metal acceleration is enabled by default
	)

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("whisper transcription failed: %w\nOutput: %s", err, string(output))
	}

	// The transcription is written to /tmp/ghospel_output.txt
	// But whisper-cli also outputs the transcription to stdout, let's parse that
	lines := strings.Split(string(output), "\n")

	var transcription strings.Builder

	// Skip header lines and extract the actual transcription
	inTranscription := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for timestamp patterns or transcription content
		if strings.Contains(line, "[00:") || inTranscription {
			inTranscription = true
			// Remove timestamp markers and extract text
			if strings.Contains(line, "]") {
				parts := strings.SplitN(line, "]", 2)
				if len(parts) > 1 {
					text := strings.TrimSpace(parts[1])
					if text != "" {
						transcription.WriteString(text)
						transcription.WriteString(" ")
					}
				}
			}
		}
	}

	result := strings.TrimSpace(transcription.String())
	if result == "" {
		// Fallback: return the full output if we couldn't parse it
		result = string(output)
	}

	return result, nil
}

// IsAvailable checks if the whisper binary is available
func (c *Client) IsAvailable() bool {
	cmd := exec.Command(c.whisperBinaryPath, "--help")
	err := cmd.Run()

	return err == nil
}
