package terminal

import (
	"fmt"
	"os"

	"github.com/AdityaKrSingh26/Glime/pkg/ansi"
	"golang.org/x/term"
)

// This package provides low level terminal control operations

// Terminal struct represnts a terminal instance state and operations
type Terminal struct {
	width          int
	height         int
	original_state *term.State
	fd             int // file description
}

// create new terminal instance
// does not enable raw mode automatically - call EnableRawMode() explicitly.
func New() (*Terminal, error) {
	fd := int(os.Stdin().Fd())

	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("stdin is not in terminal")
	}

	width, height, err := term.GetSize(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to get terminal size : %w", err)
	}

	return &Terminal{
		width:  width,
		height: height,
		fd:     fd,
	}, nil
}

// returns the current terminal width in columns.
func (t *Terminal) Width() int {
	return t.width
}

// returns the current terminal height in rows.
func (t *Terminal) Height() int {
	return t.height
}

// refresh terminal dimensions
func (t *Terminal) UpdateSize() error {
	width, height, err := term.GetSize(t.fd)
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	t.width = width
	t.height = height
	return nil
}

// clear entire screen and move terminal to top
func (t *Terminal) Clear() error {
	_, err := os.Stdout.Write([]byte(ansi.ClearScreen + ansi.MoveCursorHome))
	return err
}

// hides the cursor.
func (t *Terminal) HideCursor() error {
	_, err := os.Stdout.Write([]byte(ansi.HideCursor))
	return err
}

// allows the editor to use a separate screen
// that doesn't affect the terminal's scroll history.
func (t *Terminal) EnableAlternateBuffer() error {
	_, err := os.Stdout.Write([]byte(ansi.EnableAlternateBuffer))
	return err
}

// switches back to the main screen buffer.
func (t *Terminal) DisableAlternateBuffer() error {
	_, err := os.Stdout.Write([]byte(ansi.DisableAlternateBuffer))
	return err
}

// writes a string to the terminal.
func (t *Terminal) Write(s string) error {
	_, err := os.Stdout.WriteString(s)
	return err
}
