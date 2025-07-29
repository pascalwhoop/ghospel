# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

## Essential Commands

### Building and Development

```bash
# Build the main binary
go build -o ghospel ./cmd/ghospel

# Test the CLI
./ghospel --help
./ghospel models list
./ghospel config show

# Build for release (requires goreleaser)
goreleaser build --snapshot --clean
```

### Testing with Real Audio

```bash
# Test transcription with tiny model (fast for development)
./ghospel transcribe test-audio.mp3 --model tiny --verbose

# Test model downloads
./ghospel models download base

# Test configuration
./ghospel config set model large-v3-turbo
./ghospel config show
```

## Architecture Overview

### High-Level Design

Ghospel is a CLI audio transcription tool that wraps whisper.cpp binaries with a Go-based management
layer. The architecture follows a modular design with clear separation between CLI, business logic,
and external integrations.

**Key Design Decisions:**

- **Binary Wrapper Approach**: Uses whisper.cpp CLI binary instead of CGO bindings for simplicity
  and reliability
- **Model Caching**: Downloads and caches Whisper models from Hugging Face at `~/.whisper/`
- **FFmpeg Integration**: Converts audio formats to 16kHz mono WAV before transcription
- **Text Formatting**: Implements intelligent paragraph breaks similar to VoiceInk

### Core Components

**CLI Layer (`internal/cli/`, `internal/commands/`)**

- Built with `urfave/cli/v2` framework
- Commands: `transcribe`, `models`, `config`, `cache`
- Global flags and environment variable support

**Business Logic Layer**

- `internal/transcription/service.go`: Main transcription pipeline orchestrator
- `internal/transcription/formatter.go`: Text formatting with paragraph intelligence (50
  words/paragraph, max 4 sentences, sentence significance detection)
- `internal/models/manager.go`: Model download/cache management from Hugging Face
- `internal/config/config.go`: YAML-based configuration with defaults

**Integration Layer**

- `internal/whisper/client.go`: Wrapper for whisper.cpp binary execution
- `internal/audio/processor.go`: FFmpeg wrapper for audio format conversion
- `internal/cache/manager.go`: Cache lifecycle management

### Data Flow

1. **Input Processing**: Audio files discovered via directory traversal or direct paths
2. **Model Preparation**: Auto-download models from Hugging Face if not cached
3. **Audio Conversion**: FFmpeg converts to 16kHz mono WAV in `/tmp/ghospel`
4. **Transcription**: Execute whisper.cpp binary with Metal GPU acceleration
5. **Text Formatting**: Apply paragraph breaks using sentence analysis
6. **Output Generation**: Save formatted text alongside original audio files

### Configuration System

- Default config: `~/.config/ghospel/config.yaml`
- Environment variables: `GHOSPEL_*` prefix
- CLI flags override config values
- Model default: `large-v3-turbo` (optimal speed/quality balance)

### External Dependencies

- **whisper.cpp**: Binary transcription engine (user must install separately)
- **FFmpeg**: Audio conversion (`/opt/homebrew/bin/ffmpeg` default path)
- **Hugging Face**: Model download source (`https://huggingface.co/ggerganov/whisper.cpp`)

### Commit messages

- Use Conventional Commits

### Release Process

- auto generate release notes from changelog
- **GoReleaser**: Automated builds for macOS Intel/ARM64
- **GitHub Actions**: Triggered on git tags (`v*`)
- **Homebrew Tap**: Auto-updates `pascalwhoop/homebrew-ghospel`
- **Dependencies**: FFmpeg automatically handled via Homebrew formula

### Key Implementation Details

- **Progress Bars**: `schollz/progressbar/v3` for model downloads and processing feedback
- **Concurrent Processing**: Worker pools for batch transcription (configurable via `--workers`)
- **Error Recovery**: Graceful handling of model download failures, audio conversion issues
- **Cache Management**: Smart model caching prevents re-downloads, cleanup commands available
- **Text Intelligence**: Sentence tokenization, word counting, significance detection for readable
  output formatting
