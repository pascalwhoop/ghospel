package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pascalwhoop/ghospel/internal/config"
	"github.com/pascalwhoop/ghospel/internal/transcription"
	"github.com/urfave/cli/v2"
)

// TranscribeCommand creates the transcribe command
func TranscribeCommand() *cli.Command {
	return &cli.Command{
		Name:      "transcribe",
		Usage:     "Transcribe audio files or directories",
		ArgsUsage: "[files or directories...]",
		Description: `Transcribe audio files to text using local Whisper models.

   Supports common audio formats: MP3, M4A, WAV, FLAC, MP4, etc.
   Output files are created alongside input files with .txt extension.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "Whisper model to use (tiny, base, small, medium, large-v3, large-v3-turbo)",
				Value:   "large-v3-turbo",
				EnvVars: []string{"GHOSPEL_MODEL"},
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Custom output directory (default: same as input)",
				EnvVars: []string{"GHOSPEL_OUTPUT_DIR"},
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Usage:   "Number of concurrent workers",
				Value:   4,
				EnvVars: []string{"GHOSPEL_WORKERS"},
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Usage:   "Process directories recursively",
			},
			&cli.BoolFlag{
				Name:    "timestamps",
				Aliases: []string{"t"},
				Usage:   "Include timestamps in output",
			},
			&cli.StringFlag{
				Name:    "prompt",
				Aliases: []string{"p"},
				Usage:   "Custom transcription prompt for better accuracy",
				EnvVars: []string{"GHOSPEL_PROMPT"},
			},
			&cli.StringFlag{
				Name:    "language",
				Aliases: []string{"l"},
				Usage:   "Force specific language (default: auto-detect)",
				Value:   "auto",
				EnvVars: []string{"GHOSPEL_LANGUAGE"},
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format (txt, srt, vtt)",
				Value:   "txt",
				EnvVars: []string{"GHOSPEL_FORMAT"},
			},
			&cli.StringFlag{
				Name:    "cache-dir",
				Usage:   "Override default cache directory",
				EnvVars: []string{"GHOSPEL_CACHE_DIR"},
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress progress bars and non-error output",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"F"},
				Usage:   "Force re-transcription of files that already have output files",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.ShowCommandHelp(c, "transcribe")
			}

			// Load configuration
			cfg, err := config.Load(c.String("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override config with CLI flags
			opts := transcription.Options{
				Model:      c.String("model"),
				OutputDir:  c.String("output-dir"),
				Workers:    c.Int("workers"),
				Recursive:  c.Bool("recursive"),
				Timestamps: c.Bool("timestamps"),
				Prompt:     c.String("prompt"),
				Language:   c.String("language"),
				Format:     c.String("format"),
				CacheDir:   c.String("cache-dir"),
				Quiet:      c.Bool("quiet"),
				Verbose:    c.Bool("verbose"),
				Force:      c.Bool("force"),
			}

			// Apply config defaults
			if opts.CacheDir == "" {
				opts.CacheDir = cfg.CacheDir
			}
			if opts.Model == "large-v3-turbo" && cfg.Model != "" {
				opts.Model = cfg.Model
			}
			if opts.Workers == 4 && cfg.Workers > 0 {
				opts.Workers = cfg.Workers
			}

			// Validate output format
			validFormats := []string{"txt", "srt", "vtt"}
			formatValid := false
			for _, f := range validFormats {
				if strings.EqualFold(opts.Format, f) {
					formatValid = true
					break
				}
			}
			if !formatValid {
				return fmt.Errorf("invalid format: %s (valid: %s)", opts.Format, strings.Join(validFormats, ", "))
			}

			// Get input files/directories
			inputs := make([]string, c.NArg())
			for i := 0; i < c.NArg(); i++ {
				inputs[i], _ = filepath.Abs(c.Args().Get(i))
			}

			// Create transcription service
			service := transcription.NewService(opts)

			// Start transcription
			return service.TranscribeFiles(inputs)
		},
	}
}
