package commands

import (
	"fmt"

	"github.com/pascalwhoop/ghospel/internal/config"
	"github.com/urfave/cli/v2"
)

// ConfigCommand creates the config command
func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage configuration settings",
		Description: `View and modify ghospel configuration settings.

   Configuration is stored in ~/.config/ghospel/config.yaml`,
		Subcommands: []*cli.Command{
			{
				Name:      "show",
				Usage:     "Display current configuration",
				ArgsUsage: " ",
				Action: func(c *cli.Context) error {
					cfg, err := config.Load(c.String("config"))
					if err != nil {
						return fmt.Errorf("failed to load config: %w", err)
					}
					return config.Show(cfg)
				},
			},
			{
				Name:      "set",
				Usage:     "Set a configuration value",
				ArgsUsage: "<key> <value>",
				Description: `Set a configuration key to a specific value.

   Available keys:
     model         - Default Whisper model (tiny, base, small, medium, large, large-v3)
     cache_dir     - Directory for model and file caching  
     workers       - Number of concurrent transcription workers
     language      - Default language for transcription
     output_format - Default output format (txt, srt, vtt)
     ffmpeg_path   - Path to FFmpeg binary`,
				Action: func(c *cli.Context) error {
					if c.NArg() != 2 {
						return cli.ShowCommandHelp(c, "set")
					}

					key := c.Args().Get(0)
					value := c.Args().Get(1)

					return config.Set(c.String("config"), key, value)
				},
			},
			{
				Name:      "get",
				Usage:     "Get a configuration value",
				ArgsUsage: "<key>",
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return cli.ShowCommandHelp(c, "get")
					}

					key := c.Args().First()
					cfg, err := config.Load(c.String("config"))
					if err != nil {
						return fmt.Errorf("failed to load config: %w", err)
					}

					return config.Get(cfg, key)
				},
			},
			{
				Name:      "reset",
				Usage:     "Reset configuration to defaults",
				ArgsUsage: " ",
				Description: `Reset all configuration settings to their default values.
   
   This will overwrite your existing configuration file.`,
				Action: func(c *cli.Context) error {
					return config.Reset(c.String("config"))
				},
			},
		},
		Action: func(c *cli.Context) error {
			return cli.ShowCommandHelp(c, "config")
		},
	}
}
