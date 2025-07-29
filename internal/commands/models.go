package commands

import (
	"github.com/pascalwhoop/ghospel/internal/models"
	"github.com/urfave/cli/v2"
)

// ModelsCommand creates the models command
func ModelsCommand() *cli.Command {
	return &cli.Command{
		Name:  "models",
		Usage: "Manage Whisper models",
		Description: `Download, list, and manage Whisper models for local transcription.

   Models are cached locally and downloaded on first use.`,
		Subcommands: []*cli.Command{
			{
				Name:      "list",
				Usage:     "List available and downloaded models",
				ArgsUsage: " ",
				Action: func(c *cli.Context) error {
					manager := models.NewManager("")
					return manager.List()
				},
			},
			{
				Name:      "download",
				Usage:     "Download a specific model",
				ArgsUsage: "<model-name>",
				Description: `Download a Whisper model for offline use.

   Available models: tiny, base, small, medium, large, large-v3`,
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return cli.ShowCommandHelp(c, "download")
					}

					modelName := c.Args().First()
					manager := models.NewManager("")
					return manager.Download(modelName)
				},
			},
			{
				Name:      "cleanup",
				Usage:     "Remove unused cached models",
				ArgsUsage: " ",
				Description: `Remove old or unused model files to free up disk space.
   
   This will remove models that haven't been used recently.`,
				Action: func(c *cli.Context) error {
					manager := models.NewManager("")
					return manager.Cleanup()
				},
			},
			{
				Name:      "info",
				Usage:     "Show information about a specific model",
				ArgsUsage: "<model-name>",
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return cli.ShowCommandHelp(c, "info")
					}

					modelName := c.Args().First()
					manager := models.NewManager("")
					return manager.Info(modelName)
				},
			},
		},
		Action: func(c *cli.Context) error {
			return cli.ShowCommandHelp(c, "models")
		},
	}
}
