package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pascalwhoop/ghospel/internal/commands"
	"github.com/pascalwhoop/ghospel/internal/config"
	"github.com/urfave/cli/v2"
)

// NewApp creates a new CLI application
func NewApp() *cli.App {
	app := &cli.App{
		Name:        "ghospel",
		Usage:       "A blazing-fast, privacy-first command-line audio transcription tool for macOS",
		Description: "Ghospel transcribes audio files using local AI models optimized for Apple Silicon",
		Version:     "0.1.0",
		Authors: []*cli.Author{
			{
				Name:  "Pascal",
				Email: "pascal@example.com",
			},
		},
		Before: func(c *cli.Context) error {
			// Initialize config directory
			return config.InitConfigDir()
		},
		Commands: []*cli.Command{
			commands.TranscribeCommand(),
			commands.ModelsCommand(),
			commands.ConfigCommand(),
			commands.CacheCommand(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose output",
				EnvVars: []string{"GHOSPEL_VERBOSE"},
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to config file",
				Value:   filepath.Join(os.Getenv("HOME"), ".config", "ghospel", "config.yaml"),
				EnvVars: []string{"GHOSPEL_CONFIG"},
			},
		},
	}

	// Set custom help template
	cli.AppHelpTemplate = fmt.Sprintf(`%s
EXAMPLES:
   ghospel transcribe audio.mp3                    # Transcribe single file
   ghospel transcribe *.mp3                        # Transcribe multiple files  
   ghospel transcribe ./podcasts/ --recursive      # Transcribe directory recursively
   ghospel transcribe audio.mp3 --model large-v3   # Use specific model
   ghospel models download base                     # Download model
   ghospel config set model large-v3               # Set default model

WEBSITE: https://github.com/pascalwhoop/ghospel
`, cli.AppHelpTemplate)

	return app
}
