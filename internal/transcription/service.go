package transcription

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pascalwhoop/ghospel/internal/audio"
	"github.com/pascalwhoop/ghospel/internal/models"
	"github.com/pascalwhoop/ghospel/internal/whisper"
)

// Options holds transcription configuration
type Options struct {
	Model      string
	OutputDir  string
	Workers    int
	Recursive  bool
	Timestamps bool
	Prompt     string
	Language   string
	Format     string
	CacheDir   string
	Quiet      bool
	Verbose    bool
}

// Service handles audio transcription
type Service struct {
	opts           Options
	audioProcessor *audio.Processor
	whisperClient  *whisper.Client
	modelManager   *models.Manager
}

// NewService creates a new transcription service
func NewService(opts Options) *Service {
	// Initialize audio processor
	audioProcessor := audio.NewProcessor("/opt/homebrew/bin/ffmpeg", "/tmp/ghospel")

	// Initialize whisper client
	whisperClient := whisper.NewClient("", opts.CacheDir)

	// Initialize model manager
	modelManager := models.NewManager(opts.CacheDir)

	return &Service{
		opts:           opts,
		audioProcessor: audioProcessor,
		whisperClient:  whisperClient,
		modelManager:   modelManager,
	}
}

// TranscribeFiles transcribes the given input files/directories
func (s *Service) TranscribeFiles(inputs []string) error {
	if !s.opts.Quiet {
		fmt.Printf("🎵 Ghospel v0.1.0 - Starting transcription with model: %s\n", s.opts.Model)
	}

	// Find all audio files
	audioFiles, err := s.findAudioFiles(inputs)
	if err != nil {
		return fmt.Errorf("failed to find audio files: %w", err)
	}

	if len(audioFiles) == 0 {
		return fmt.Errorf("no audio files found")
	}

	if !s.opts.Quiet {
		fmt.Printf("📁 Found %d audio file(s) to transcribe\n", len(audioFiles))
	}

	// TODO: Implement actual transcription pipeline
	// For now, just create placeholder output files
	for _, file := range audioFiles {
		if err := s.transcribeFile(file); err != nil {
			if s.opts.Verbose {
				fmt.Printf("❌ Failed to transcribe %s: %v\n", file, err)
			}

			continue
		}

		if !s.opts.Quiet {
			fmt.Printf("✅ Transcribed: %s\n", filepath.Base(file))
		}
	}

	if !s.opts.Quiet {
		fmt.Println("🎉 Transcription complete!")
	}

	return nil
}

// findAudioFiles discovers audio files from the input paths
func (s *Service) findAudioFiles(inputs []string) ([]string, error) {
	var audioFiles []string

	supportedExts := []string{".mp3", ".m4a", ".wav", ".flac", ".mp4", ".aac", ".ogg"}

	for _, input := range inputs {
		stat, err := os.Stat(input)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", input, err)
		}

		if stat.IsDir() {
			// Handle directory
			if s.opts.Recursive {
				err = filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if !info.IsDir() && s.isAudioFile(path, supportedExts) {
						audioFiles = append(audioFiles, path)
					}

					return nil
				})
			} else {
				entries, err := os.ReadDir(input)
				if err != nil {
					return nil, fmt.Errorf("cannot read directory %s: %w", input, err)
				}

				for _, entry := range entries {
					if !entry.IsDir() {
						path := filepath.Join(input, entry.Name())
						if s.isAudioFile(path, supportedExts) {
							audioFiles = append(audioFiles, path)
						}
					}
				}
			}

			if err != nil {
				return nil, err
			}
		} else {
			// Handle file
			if s.isAudioFile(input, supportedExts) {
				audioFiles = append(audioFiles, input)
			}
		}
	}

	return audioFiles, nil
}

// isAudioFile checks if the file has a supported audio extension
func (s *Service) isAudioFile(path string, supportedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}

	return false
}

// transcribeFile transcribes a single audio file
func (s *Service) transcribeFile(inputPath string) error {
	// Determine output file path
	outputPath := s.getOutputPath(inputPath)

	// Step 1: Check if model is downloaded, download if needed
	if err := s.ensureModelDownloaded(); err != nil {
		return fmt.Errorf("model preparation failed: %w", err)
	}

	// Step 2: Convert audio to WAV using FFmpeg if needed
	wavPath, needsCleanup, err := s.prepareAudioFile(inputPath)
	if err != nil {
		return fmt.Errorf("audio preparation failed: %w", err)
	}

	// Clean up temporary WAV file if needed
	if needsCleanup {
		defer s.audioProcessor.Cleanup(wavPath)
	}

	// Step 3: Run Whisper inference
	transcription, err := s.whisperClient.Transcribe(wavPath, s.opts.Model)
	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}

	// Step 4: Format and save output
	content := s.formatOutput(transcription, inputPath)
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ensureModelDownloaded checks if the model exists and downloads it if needed
func (s *Service) ensureModelDownloaded() error {
	availableModels := s.modelManager.AvailableModels()

	var targetModel *models.ModelInfo

	for i, model := range availableModels {
		if model.Name == s.opts.Model {
			targetModel = &availableModels[i]
			break
		}
	}

	if targetModel == nil {
		return fmt.Errorf("unknown model: %s", s.opts.Model)
	}

	// Check if model file exists
	if _, err := os.Stat(targetModel.Path); os.IsNotExist(err) {
		if !s.opts.Quiet {
			fmt.Printf("📥 Model %s not found, downloading...\n", s.opts.Model)
		}

		return s.modelManager.Download(s.opts.Model)
	}

	return nil
}

// prepareAudioFile converts audio to WAV format if needed
func (s *Service) prepareAudioFile(inputPath string) (string, bool, error) {
	// Check if file is already in WAV format
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext == ".wav" {
		// TODO: Check if it's 16kHz mono, if not, still convert
		return inputPath, false, nil
	}

	// Convert to WAV
	if !s.opts.Quiet && s.opts.Verbose {
		fmt.Printf("🔄 Converting %s to WAV format...\n", filepath.Base(inputPath))
	}

	wavPath, err := s.audioProcessor.ConvertToWav(inputPath)
	if err != nil {
		return "", false, err
	}

	return wavPath, true, nil
}

// formatOutput formats the transcription output
func (s *Service) formatOutput(transcription, inputPath string) string {
	var content strings.Builder

	// Add header comment
	content.WriteString(fmt.Sprintf("# Transcription of: %s\n", filepath.Base(inputPath)))
	content.WriteString(fmt.Sprintf("# Model: %s\n", s.opts.Model))
	content.WriteString("# Generated with Ghospel v0.1.0\n\n")

	// Format the transcription into readable paragraphs
	formatter := NewTextFormatter()
	formattedText := formatter.Format(transcription)

	// Add the formatted transcription
	content.WriteString(formattedText)
	content.WriteString("\n")

	return content.String()
}

// getOutputPath determines the output file path
func (s *Service) getOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	if s.opts.OutputDir != "" {
		dir = s.opts.OutputDir
		// Ensure output directory exists
		os.MkdirAll(dir, 0o755)
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := "." + s.opts.Format

	return filepath.Join(dir, base+ext)
}
