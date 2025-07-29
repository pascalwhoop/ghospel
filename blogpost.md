+++
title = 'Writing Ghospel, a golang based whisper.cpp wrapper'
date = 2025-07-29T11:17:14+02:00
draft = false
author = 'Pascal'
+++

{{< fimg ghospel Crop "1920x1080" />}}

## Motivation

At my day job we use a lot of language models and AI-assisted programming for developing various
products. Since all our teams wanted to align on our ways of working with language models, I decided
to research how the best in the industry use LLMs.

This led me down the rabbit hole of discovering the AI Engineer community—an extremely high quality
collection of people and presentations. Their YouTube channel contains many excellent talks that
seemed like exactly what I needed to understand how others work with language models.

Which in turn led me to wanting transcripts for all these videos. I've been curious about having a
tool that can transcribe large collections of audio files—imagine transcribing all episodes of a
specific podcaster. So I decided to build something for myself.

I know that whisper.cpp is the go-to for audio transcription on Mac. I extensively use VoiceInk to
transcribe my speech—in fact, I'm dictating this very article because it's much faster for me to
talk out blog posts than write them.

But most existing tools didn't have the UX I wanted: just point a CLI at a folder full of audio
files and have it figure out how to transcribe them all into text.

GUI applications usually fall short on scriptability. You can drag individual files, but you can't
easily drag a whole folder and get a complete set of transcripts back. And while whisper.cpp is
powerful, it's not very user-friendly for batch operations.

Since I've always wanted to write Golang but never found the right opportunity to learn it, this
seemed perfect: a CLI in Go that wraps whisper.cpp, handles automatic model downloading, and
provides a nice UX for batch transcription.

And so Ghospel was born—a Golang whisper wrapper.

## Approach

I drew inspiration from two key projects:

- [VoiceInk](https://github.com/Beingpax/VoiceInk) - for its elegant model auto-downloading UX
- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) - the core transcription engine

I decided to take Claude Code for a spin. First, I set up the project structure:

```bash
git init
mkdir inspiration/ && cd inspiration
git clone https://github.com/ggerganov/whisper.cpp
git clone https://github.com/Beingpax/VoiceInk
```

I asked Claude to build a CLI that could handle a folder full of mp3 files and transcribe them into
text, taking heavy inspiration from both libraries. VoiceInk's automatic model downloading was
particularly elegant, while whisper.cpp would be the heavy lifter for actual transcription.

After describing my requirements, Claude generated
[the README.md](https://github.com/pascalwhoop/ghospel/blob/main/README.md) in a 'target state'
format—as if the library was already finished. I've found this approach works exceptionally well
with LLMs because you're describing a target state and the LLM works backwards from the UX to the
implementation.

## Building Ghospel: 4 Hours Start to Finish

What followed was one of the most productive coding sessions I've had in years. In about 4 hours
total, we went from concept to a production-ready CLI tool.

### Foundation (Hour 1)

Claude started by creating a well-structured Go project with clean separation of concerns:

- **CLI layer** using `urfave/cli/v2` for command parsing and flags
- **Business logic** layer with separate services for transcription, models, and audio processing
- **Integration layer** wrapping whisper.cpp binaries and FFmpeg

Rather than building a monolithic script, we architected a proper Go application with clear
interfaces between components from day one.

### Core Features (Hour 2)

With the foundation solid, we rapidly implemented the essential functionality:

- **Model management**: Automatic downloading from Hugging Face with smart caching
- **Audio processing**: FFmpeg integration for format conversion to 16kHz mono WAV
- **Batch processing**: Parallel transcription with configurable worker pools
- **Text formatting**: Intelligent paragraph breaks similar to VoiceInk's approach

I specifically asked for progress bars during long operations, and Claude implemented comprehensive
progress tracking with detailed statistics.

### Production Polish (Hour 3-4)

The final stretch focused on making Ghospel production-ready:

- **Release automation**: Complete GitHub Actions workflow with GoReleaser
- **Cross-platform binaries**: Supporting macOS Intel and Apple Silicon
- **Error handling**: Graceful failures with detailed error messages
- **Configuration system**: YAML config with environment variable overrides
- **Documentation**: README, contributing guidelines, and usage examples

## Last Enhancement: Resumable Transcription

A few hours later, I realized Ghospel needed one more critical feature. When transcribing hundreds
of audio files, crashes are inevitable—restarting from scratch is painful. In fact it crashed
because I moved the whole project folder. So I added a flag to force re-transcription.

This led to another productive 10 minutes long session where we implemented:

- **File skip logic**: Automatically skip already transcribed files
- **Force flag**: `--force` option to override and re-transcribe existing files
- **Smart progress tracking**: Shows "Found 103 files, 33 already transcribed, 70 to process"
- **Enhanced release automation**: Conventional commits with structured changelog generation

```bash
make clean
make release
```

and when I re-ran the command, it picked up where it left off. Lovely!

A quick "please release this as v0.1.1 patch" and
[we were done](https://github.com/pascalwhoop/ghospel/releases/tag/v0.1.1).

In

## Technical Deep Dive

Looking at the final codebase, we wrote approximately 2,100 lines of Go code across 14 modules:

```
     426 internal/transcription/service.go    # Main orchestration logic
     281 internal/models/manager.go           # Model downloading & caching
     224 internal/config/config.go            # Configuration management
     206 internal/cache/manager.go            # Cache lifecycle
     169 internal/transcription/formatter.go  # Text formatting intelligence
     154 internal/commands/transcribe.go      # CLI command implementation
     127 internal/audio/processor.go          # FFmpeg wrapper
     125 internal/whisper/client.go           # whisper.cpp binary interface
```

The architecture follows clean separation of concerns with clear interfaces between layers. Each
component has a single responsibility and can be tested in isolation.

### Key Design Decisions

**Binary Wrapper Approach**: Rather than using CGO bindings, we wrap the whisper.cpp CLI binary.
This gives us simplicity and reliability—no complex build dependencies or linking issues.

**Model Caching Strategy**: Models are downloaded once from Hugging Face and cached at
`~/.whisper/`. The system is smart about avoiding re-downloads and handles interrupted downloads
gracefully.

**Intelligent Text Formatting**: The formatter implements paragraph intelligence similar to
VoiceInk. 50 words per paragraph, maximum 4 sentences, and sentence significance detection for
readable output.

**Concurrent Processing**: Worker pools allow parallel transcription of multiple files, with the
concurrency level configurable based on system resources.

## Results

After about 4 hours of development, Ghospel delivers exactly what I wanted:

- **Just works**: Point it at a folder, get transcripts back
- **Production ready**: Proper error handling, logging, and progress tracking
- **Apple Silicon optimized**: Metal GPU acceleration on M-series chips
- **Resumable**: Skip already processed files for interrupted batch jobs
- **Maintainable**: Clean architecture with comprehensive documentation

More importantly, it demonstrates the current state of AI-assisted development. Working with Claude
Code was like pair programming with a very capable engineer who understood both technical
requirements and user experience goals. Honestly, I felt more like a product manager than a
developer at this point

The tool handled architecture decisions, implemented Go best practices I wasn't familiar with, wrote
documentation, set up CI/CD, and even generated proper release notes.

Meanwhile, I am calling it an evening and letting this thing finish transcribing the rest of the
files over night.

```bash
  ~/Code/business/ghospel  main ?1 ❯ ./ghospel transcribe --output-dir outputs/ data/*.mp3
🎵 Ghospel v0.1.0 - Starting transcription with model: large-v3-turbo
📁 Found 452 audio file(s), 9 already transcribed, 443 to process
```

This is one of these moments where I regret having traded in my macbook pro for a macbook air
without a fan. It really takes a good long while longer to transcribe all these files. But 2-3 night
shifts should do the trick here.

## Looking Forward

The implementation really went crazy fast. Maybe it is because golang is so much more structured
that mistakes are told to you by a great compiler. Or maybe claude code is just that good. I do not
know yet.

Now I can get back to the original goal from earlier today, writing our ways of working with LLMs.
And certainly claude-code / AI assisted coding is going to be added to our must-have things to do.

I also feel like there are a lot of moments when things are still not quite how I would like them to
be and I'd love to have a sort of 'team wide memory'. Because I want corrections I make to be
propagated to the rest of the team so that we all adhere to this standard going forward.

_Ghospel is [open source on GitHub](https://github.com/pascalwhoop/ghospel). The entire development
history is preserved in the commit log, showing exactly how Claude Code and I built this together._
