package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Processor handles audio file processing and conversion
type Processor struct {
	ffmpegPath string
	tempDir    string
}

// NewProcessor creates a new audio processor
func NewProcessor(ffmpegPath, tempDir string) *Processor {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg" // Default to system ffmpeg
	}

	if tempDir == "" {
		tempDir = "/tmp/ghospel"
	}

	// Ensure temp directory exists
	os.MkdirAll(tempDir, 0o755)

	return &Processor{
		ffmpegPath: ffmpegPath,
		tempDir:    tempDir,
	}
}

// ConvertToWav converts an audio file to 16kHz mono WAV format required by Whisper
func (p *Processor) ConvertToWav(inputPath string) (string, error) {
	// Generate output filename
	inputBase := filepath.Base(inputPath)
	inputExt := filepath.Ext(inputBase)
	outputName := strings.TrimSuffix(inputBase, inputExt) + "_converted.wav"
	outputPath := filepath.Join(p.tempDir, outputName)

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// FFmpeg command to convert to 16kHz mono WAV
	cmd := exec.Command(p.ffmpegPath,
		"-i", inputPath, // Input file
		"-ar", "16000", // Sample rate: 16kHz (required by Whisper)
		"-ac", "1", // Audio channels: 1 (mono)
		"-c:a", "pcm_s16le", // Audio codec: 16-bit PCM
		"-f", "wav", // Output format: WAV
		"-y",       // Overwrite output file
		outputPath, // Output file
	)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	// Verify the output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output file was not created: %s", outputPath)
	}

	return outputPath, nil
}

// GetAudioInfo returns basic information about an audio file
func (p *Processor) GetAudioInfo(inputPath string) (map[string]string, error) {
	cmd := exec.Command(p.ffmpegPath,
		"-i", inputPath,
		"-hide_banner",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		// ffmpeg returns non-zero exit code when using -f null, but still provides info
		// So we ignore the error and parse the output
	}

	info := make(map[string]string)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Duration:") {
			// Extract duration
			parts := strings.Split(line, ",")
			if len(parts) > 0 {
				duration := strings.TrimSpace(strings.Replace(parts[0], "Duration:", "", 1))
				info["duration"] = duration
			}
		}

		if strings.Contains(line, "Audio:") {
			// Extract audio format info
			info["audio_info"] = line
		}
	}

	return info, nil
}

// Cleanup removes temporary files
func (p *Processor) Cleanup(filePath string) error {
	if strings.Contains(filePath, p.tempDir) {
		return os.Remove(filePath)
	}

	return nil
}

// IsFFmpegAvailable checks if FFmpeg is available on the system
func (p *Processor) IsFFmpegAvailable() bool {
	cmd := exec.Command(p.ffmpegPath, "-version")
	err := cmd.Run()

	return err == nil
}
