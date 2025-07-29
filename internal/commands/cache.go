package commands

import (
	"github.com/pascalwhoop/ghospel/internal/cache"
	"github.com/urfave/cli/v2"
)

// CacheCommand creates the cache command
func CacheCommand() *cli.Command {
	return &cli.Command{
		Name:  "cache",
		Usage: "Manage download and processing cache",
		Description: `Manage cached files including models, downloaded audio, and temporary files.

   Cache is stored in ~/.whisper/ by default.`,
		Subcommands: []*cli.Command{
			{
				Name:      "info",
				Usage:     "Show cache statistics",
				ArgsUsage: " ",
				Description: `Display information about cache usage including:
   - Total cache size
   - Number of cached files
   - Cache directory location
   - Last cleanup date`,
				Action: func(c *cli.Context) error {
					manager := cache.NewManager("")
					return manager.Info()
				},
			},
			{
				Name:      "clean",
				Usage:     "Remove old cached files",
				ArgsUsage: " ",
				Description: `Remove cached files older than the retention period.
   
   This preserves recently used models and files while cleaning up old data.`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "older-than",
						Usage: "Remove files older than duration (e.g., 30d, 7d, 24h)",
						Value: "30d",
					},
				},
				Action: func(c *cli.Context) error {
					manager := cache.NewManager("")
					return manager.Clean(c.String("older-than"))
				},
			},
			{
				Name:      "clear",
				Usage:     "Clear entire cache",
				ArgsUsage: " ",
				Description: `Remove all cached files including models and temporary files.
   
   WARNING: This will require re-downloading models on next use.`,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Skip confirmation prompt",
					},
				},
				Action: func(c *cli.Context) error {
					manager := cache.NewManager("")
					return manager.Clear(c.Bool("force"))
				},
			},
			{
				Name:      "path",
				Usage:     "Show cache directory path",
				ArgsUsage: " ",
				Action: func(c *cli.Context) error {
					manager := cache.NewManager("")
					return manager.ShowPath()
				},
			},
		},
		Action: func(c *cli.Context) error {
			return cli.ShowCommandHelp(c, "cache")
		},
	}
}
