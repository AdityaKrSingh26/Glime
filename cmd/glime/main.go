package main

import (
	"fmt"
	"os"

	"github.com/AdityaKrSingh26/Glime/internal/editor"
)

const version = "0.1.0"

func main() {
	// Parse command line arguments
	args := os.Args[1:]

	// Handle flags
	if len(args) > 0 {
		switch args[0] {
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			fmt.Printf("Glime version %s\n", version)
			return
		}
	}

	// Create editor
	ed, err := editor.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating editor: %v\n", err)
		os.Exit(1)
	}

	// Load file if specified
	if len(args) > 0 {
		filePath := args[0]
		if err := ed.LoadFile(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
			os.Exit(1)
		}
	}

	// Run editor
	if err := ed.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	help := `Glime - A terminal-based modal text editor

Usage:
  glime [options] [file]

Options:
  -h, --help      Show this help message
  -v, --version   Show version information

Examples:
  glime                 Open with empty buffer
  glime file.txt        Open or create file.txt
  glime /path/to/file   Open file at path

Key Bindings:
  Normal Mode:
    i          Enter insert mode
    :          Enter command mode
    h,j,k,l    Move cursor left/down/up/right
    Arrow keys Move cursor
    0          Move to line start
    $          Move to line end
    g          Move to first line
    G          Move to last line
    x          Delete character

  Insert Mode:
    ESC        Return to normal mode
    Backspace  Delete previous character
    Enter      New line

  Command Mode:
    :w         Write (save) file
    :q         Quit
    :wq        Write and quit
    :q!        Force quit (discard changes)
    :{number}  Go to line number

Report bugs at: https://github.com/AdityaKrSingh26/glime/issues
`
	fmt.Print(help)
}
