# Glime

A terminal-based modal text editor written from scratch in Go. Glime is a lightweight, terminal-based text editor built entirely in Go, inspired by Vim's modal editing philosophy. This project demonstrates deep systems programming knowledge, clean architecture, and professional Go development practices.


## Features

- **Modal Editing** - Vim-inspired Normal, Insert, and Command modes
- **Fast & Lightweight** - Minimal dependencies, pure Go implementation
- **Core Editing** - Insert, delete, navigate, save, and load files
- **Full Keyboard Support** - Arrow keys, special keys, UTF-8 characters
- **Status Bar** - Real-time mode, file, and cursor information
- **Clean Architecture** - Well-organized, testable, maintainable code

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/glime.git
cd glime

# Build the binary
make build

# Or install to system
make install
```

### Usage

```bash
# Open Glime
./bin/glime

# Open a specific file
./bin/glime filename.txt

# Show help
./bin/glime --help
```

## Key Bindings

### Normal Mode

| Key | Action |
|-----|--------|
| `i` | Enter insert mode |
| `:` | Enter command mode |
| `h` `j` `k` `l` | Move cursor left/down/up/right |
| `Arrow keys` | Move cursor |
| `0` | Move to line start |
| `$` | Move to line end |
| `gg` | Go to first line |
| `G` | Go to last line |
| `x` | Delete character |
| `dd` | Delete line |

### Insert Mode

| Key | Action |
|-----|--------|
| `ESC` | Return to normal mode |
| `Backspace` | Delete previous character |
| `Enter` | New line |
| Any character | Insert character |

### Command Mode

| Command | Action |
|---------|--------|
| `:w` | Save file |
| `:q` | Quit editor |
| `:wq` | Save and quit |
| `:q!` | Quit without saving |

## Project Structure

```
glime/
├── cmd/glime/          # Main entry point
├── internal/           # Internal packages
│   ├── buffer/        # Text buffer implementation
│   ├── cursor/        # Cursor management
│   ├── editor/        # Main editor logic
│   ├── input/         # Keyboard input handling
│   ├── terminal/      # Terminal control
│   ├── ui/            # UI rendering
│   └── file/          # File I/O
├── pkg/               # Public packages
├── docs/              # Documentation
└── examples/          # Example files
```

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for build automation)

### Building from Source

```bash
# Install dependencies
make deps

# Run tests
make test

# Run with coverage
make coverage

# Run linter
make lint

# Format code
make fmt

# Run all checks
make pre-commit
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## Architecture

Glime follows clean architecture principles with clear separation of concerns:

- **Terminal Layer**: Low-level terminal operations (raw mode, ANSI codes)
- **Buffer Layer**: Text storage and manipulation
- **Editor Layer**: State machine and command orchestration
- **UI Layer**: Rendering and display logic