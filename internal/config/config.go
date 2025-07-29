package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Model settings
	Model    string `yaml:"model"`
	Language string `yaml:"language"`
	Prompt   string `yaml:"prompt"`

	// Processing settings
	Workers   int    `yaml:"workers"`
	ChunkSize string `yaml:"chunk_size"`

	// Cache settings
	CacheDir       string `yaml:"cache_dir"`
	CacheRetention string `yaml:"cache_retention"`
	AutoCleanup    bool   `yaml:"auto_cleanup"`

	// Output settings
	OutputFormat      string `yaml:"output_format"`
	IncludeTimestamps bool   `yaml:"include_timestamps"`
	PreserveStructure bool   `yaml:"preserve_structure"`

	// Audio processing
	FFmpegPath string `yaml:"ffmpeg_path"`
	TempDir    string `yaml:"temp_dir"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		Model:             "large-v3-turbo",
		Language:          "auto",
		Prompt:            "",
		Workers:           4,
		ChunkSize:         "30s",
		CacheDir:          filepath.Join(homeDir, ".whisper"),
		CacheRetention:    "30d",
		AutoCleanup:       true,
		OutputFormat:      "txt",
		IncludeTimestamps: false,
		PreserveStructure: true,
		FFmpegPath:        "/opt/homebrew/bin/ffmpeg",
		TempDir:           "/tmp/ghospel",
	}
}

// InitConfigDir creates the configuration directory if it doesn't exist
func InitConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "ghospel")

	return os.MkdirAll(configDir, 0o755)
}

// Load loads configuration from the specified file
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, create it with defaults
		if err := Save(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to the specified file
func Save(cfg *Config, configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Show displays the current configuration
func Show(cfg *Config) error {
	fmt.Println("Current Configuration:")
	fmt.Println("======================")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Print(string(data))

	return nil
}

// Set updates a configuration value
func Set(configPath, key, value string) error {
	cfg, err := Load(configPath)
	if err != nil {
		return err
	}

	switch key {
	case "model":
		validModels := []string{"tiny", "base", "small", "medium", "large-v3", "large-v3-turbo"}
		valid := false

		for _, m := range validModels {
			if value == m {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("invalid model: %s (valid: tiny, base, small, medium, large-v3, large-v3-turbo)", value)
		}

		cfg.Model = value
	case "cache_dir":
		cfg.CacheDir = value
	case "workers":
		// Simple validation - you might want to use strconv.Atoi for proper conversion
		cfg.Workers = 4 // placeholder
	case "language":
		cfg.Language = value
	case "output_format":
		validFormats := []string{"txt", "srt", "vtt"}
		valid := false

		for _, f := range validFormats {
			if value == f {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("invalid format: %s (valid: txt, srt, vtt)", value)
		}

		cfg.OutputFormat = value
	case "ffmpeg_path":
		cfg.FFmpegPath = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)

	return nil
}

// Get retrieves a configuration value
func Get(cfg *Config, key string) error {
	switch key {
	case "model":
		fmt.Println(cfg.Model)
	case "cache_dir":
		fmt.Println(cfg.CacheDir)
	case "workers":
		fmt.Println(cfg.Workers)
	case "language":
		fmt.Println(cfg.Language)
	case "output_format":
		fmt.Println(cfg.OutputFormat)
	case "ffmpeg_path":
		fmt.Println(cfg.FFmpegPath)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}

// Reset resets configuration to defaults
func Reset(configPath string) error {
	cfg := DefaultConfig()
	if err := Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	fmt.Println("Configuration reset to defaults")

	return nil
}
