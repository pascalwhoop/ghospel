# Contributing to Ghospel

Thank you for your interest in contributing to Ghospel! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Development Guidelines](#development-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

- **macOS 12.0+** (Monterey or later)
- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **FFmpeg** - Install via `brew install ffmpeg`
- **Git** - For version control
- **whisper.cpp** - For local development (optional, can use binary)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ghospel.git
   cd ghospel
   ```
3. Add the original repository as upstream:
   ```bash
   git remote add upstream https://github.com/pascalwhoop/ghospel.git
   ```

## Development Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod tidy

# Install development tools (optional)
go install github.com/goreleaser/goreleaser@latest
```

### 2. Development Setup

Use the Makefile for automated setup:

```bash
# Complete development setup (initializes submodule + builds whisper.cpp)
make dev-setup

# Or step by step:
make build-whisper    # Build whisper.cpp with Metal optimization
make build           # Build the Go binary
make test            # Run tests
```

### 3. Manual Build (Alternative)

```bash
# Build the binary
go build -o ghospel ./cmd/ghospel

# Test the build
./ghospel --version
./ghospel --help
```

### 4. Set Up whisper.cpp (Manual - if not using Makefile)

The project uses whisper.cpp as a Git submodule:

```bash
# Initialize and update submodule
git submodule update --init --recursive

# Build with Metal support (macOS)
cd whisper_cpp_source
cmake -B build -DGGML_METAL=ON -DGGML_METAL_EMBED_LIBRARY=ON -DGGML_BLAS_DEFAULT=ON \
    -DCMAKE_BUILD_TYPE=Release -DWHISPER_BUILD_TESTS=OFF -DWHISPER_BUILD_SERVER=OFF
cmake --build build -j --config Release

# Test whisper-cli binary
./build/bin/whisper-cli --help
```

### 5. Development Environment

```bash
# Set up development cache directory
mkdir -p ~/.whisper-dev

# Export development environment variables
export GHOSPEL_CACHE_DIR="$HOME/.whisper-dev"
export GHOSPEL_LOG_LEVEL="debug"
export GHOSPEL_VERBOSE="true"
```

## Contributing Process

### 1. Choose an Issue

- Look for issues labeled `good first issue` for beginners
- Check if an issue is already assigned before starting work
- Comment on the issue to let others know you're working on it

### 2. Create a Branch

```bash
# Stay up to date with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 3. Make Changes

- Write clean, well-documented code
- Follow existing code style and conventions
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes

```bash
# Run code quality checks
make lint              # Run linters (catches unused imports, etc.)
make fmt               # Format code
make vet               # Run go vet

# Build and test functionality
make build             # Build the binary
./ghospel models list
./ghospel config show

# Test with real audio files
./ghospel transcribe test-audio.mp3 --model tiny

# Test different formats and options
./ghospel transcribe test-audio.m4a --model base --verbose
```

### 5. Commit and Push

```bash
# Stage your changes
git add .

# Commit with descriptive message
git commit -m "feat: add support for custom prompts

- Add --prompt flag to transcribe command
- Update configuration to store default prompt
- Add validation for prompt length
- Update documentation and help text"

# Push to your fork
git push origin feature/your-feature-name
```

### 6. Create a Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Fill out the PR template with:
   - Clear description of changes
   - Link to related issues
   - Testing instructions
   - Screenshots/examples if applicable

## Development Guidelines

### Code Style

- **Linting**: Run `make lint` before committing (catches unused imports, formatting issues, etc.)
- **Formatting**: Use `make fmt` to auto-format code with gofmt and goimports
- **Imports**: Group imports (standard, external, internal) - handled automatically by goimports
- **Naming**: Use clear, descriptive names
- **Comments**: Document public functions and complex logic
- **Error handling**: Always handle errors appropriately

#### Automated Code Quality

We use **golangci-lint** (Go's equivalent of Python's ruff) to catch:
- Unused imports and variables
- Formatting issues
- Potential bugs
- Security issues
- Code style violations

```bash
make lint       # Run all linters
make lint-fix   # Auto-fix issues where possible
make fmt        # Format code
make vet        # Run go vet
```

```go
// Good
func (m *Manager) Download(modelName string) error {
    // Validate model name
    targetModel := m.findModel(modelName)
    if targetModel == nil {
        return fmt.Errorf("unknown model: %s", modelName)
    }
    
    // Check if already downloaded
    if m.isDownloaded(targetModel) {
        return nil
    }
    
    return m.downloadModel(targetModel)
}

// Bad
func (m *Manager) Download(n string) error {
    t := m.find(n)
    if t == nil { return errors.New("bad model") }
    if m.check(t) { return nil }
    return m.dl(t)
}
```

### Project Structure

```
ghospel/
├── cmd/ghospel/           # Main CLI entry point
│   └── main.go
├── internal/              # Internal packages
│   ├── audio/            # Audio processing (FFmpeg)
│   ├── cache/            # Cache management
│   ├── cli/              # CLI app configuration
│   ├── commands/         # CLI command implementations
│   ├── config/           # Configuration management
│   ├── models/           # Model download/management
│   ├── transcription/    # Core transcription pipeline
│   └── whisper/          # Whisper client wrapper
├── docs/                 # Documentation
├── examples/             # Usage examples
└── scripts/              # Build/development scripts
```

### Adding New Features

#### 1. New CLI Command

```go
// 1. Add command to internal/commands/
func NewFeatureCommand() *cli.Command {
    return &cli.Command{
        Name:  "newfeature",
        Usage: "Description of new feature",
        Flags: []cli.Flag{
            // Add flags here
        },
        Action: func(c *cli.Context) error {
            // Implementation here
            return nil
        },
    }
}

// 2. Register in internal/cli/app.go
Commands: []*cli.Command{
    commands.TranscribeCommand(),
    commands.ModelsCommand(),
    commands.ConfigCommand(),
    commands.CacheCommand(),
    commands.NewFeatureCommand(), // Add here
},
```

#### 2. New Configuration Option

```go
// 1. Add to Config struct in internal/config/config.go
type Config struct {
    // Existing fields...
    NewOption string `yaml:"new_option"`
}

// 2. Add to DefaultConfig()
func DefaultConfig() *Config {
    return &Config{
        // Existing defaults...
        NewOption: "default_value",
    }
}

// 3. Add to Set() function validation
case "new_option":
    // Add validation logic
    cfg.NewOption = value
```

#### 3. New Model Support

```go
// Add to AvailableModels() in internal/models/manager.go
{
    Name:        "new-model",
    Size:        "X.X GB",
    Description: "Description of new model",
    Path:        filepath.Join(m.cacheDir, "ggml-new-model.bin"),
    DownloadURL: fmt.Sprintf("%s/ggml-new-model.bin", baseURL),
},
```

## Testing

### Manual Testing

```bash
# Test basic commands
./ghospel --help
./ghospel transcribe --help
./ghospel models list
./ghospel config show

# Test with different audio formats
./ghospel transcribe sample.mp3 --model tiny
./ghospel transcribe sample.m4a --model base --verbose
./ghospel transcribe sample.wav --model small --timestamps

# Test batch processing
./ghospel transcribe ./audio-folder/ --recursive --model tiny

# Test configuration
./ghospel config set model large-v3-turbo
./ghospel config show
./ghospel config reset
```

### Test Cases to Cover

- [ ] **Audio formats**: MP3, M4A, WAV, FLAC, MP4
- [ ] **Models**: All supported models download and work
- [ ] **Batch processing**: Directory processing with/without recursive
- [ ] **Configuration**: All config options work correctly
- [ ] **Error handling**: Graceful failure for invalid inputs
- [ ] **Performance**: Large files and batch processing
- [ ] **Cache management**: Models cache correctly, no re-downloads

### Performance Testing

```bash
# Test with large files
./ghospel transcribe large-podcast.mp3 --model large-v3-turbo --verbose

# Test batch processing
time ./ghospel transcribe ./large-audio-collection/ --recursive --workers 4

# Test model download performance
rm ~/.whisper/ggml-base.bin
time ./ghospel transcribe sample.mp3 --model base
```

## Documentation

### Code Documentation

- Document all public functions and methods
- Add inline comments for complex logic
- Use clear variable and function names
- Include examples in function documentation

```go
// Download downloads a Whisper model from Hugging Face.
// It returns early if the model is already cached locally.
// 
// Example:
//   err := manager.Download("large-v3-turbo")
//   if err != nil {
//       log.Fatal(err)
//   }
func (m *Manager) Download(modelName string) error {
    // Implementation...
}
```

### User Documentation

- Update README.md for new features
- Add examples to documentation
- Update help text and usage information
- Add troubleshooting info for common issues

### Commit Messages

Use conventional commit format:

```
type(scope): description

- Detailed explanation of changes
- Why the change was made
- Any breaking changes

Fixes #123
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

## Release Process

### Development vs Release Builds

**Development builds**: Use local whisper.cpp binary from `./whisper_cpp_source/build/bin/whisper-cli`

**Release builds**: Embed whisper.cpp binaries directly in the Go executable

### Creating Releases

Releases are automated via GitHub Actions when tags are pushed:

```bash
# Create and push a new version tag
git tag v1.2.3
git push origin v1.2.3
```

This triggers:
1. **Submodule setup**: Initializes whisper.cpp submodule
2. **Binary builds**: Builds whisper.cpp for all target architectures with optimizations
3. **Embedding**: Embeds binaries using Go's embed feature
4. **Release**: GitHub release with self-contained binaries
5. **Homebrew**: Updates the Homebrew tap

### Manual Release Build

```bash
# Build release version with embedded binaries
make release

# Or use the comprehensive script
./scripts/build-release.sh
```

### Version Numbering

We use semantic versioning (SemVer):
- `v1.0.0`: Major version (breaking changes)
- `v1.1.0`: Minor version (new features)
- `v1.1.1`: Patch version (bug fixes)

## Getting Help

### Communication

- **GitHub Issues**: For bugs, feature requests, and questions
- **GitHub Discussions**: For general questions and community chat
- **Pull Request Reviews**: For code-specific discussions

### Common Development Issues

**Build fails with import errors:**
```bash
go mod tidy
go clean -cache
```

**whisper.cpp build issues:**
```bash
cd whisper_cpp_source
make clean
WHISPER_METAL=1 make -j
```

**Permission errors with cache:**
```bash
sudo chown -R $USER ~/.whisper
chmod -R 755 ~/.whisper
```

**FFmpeg not found:**
```bash
brew install ffmpeg
# Or specify custom path in config
```

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- README acknowledgments section

Thank you for contributing to Ghospel! 🎉

---

**Questions?** Open an issue or start a discussion on GitHub.
