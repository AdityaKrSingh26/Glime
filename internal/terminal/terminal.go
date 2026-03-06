package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	builder        strings.Builder
	input          *inputReader
}

// create new terminal instance
// does not enable raw mode automatically - call EnableRawMode() explicitly.
func New() (*Terminal, error) {
	fd := int(os.Stdin.Fd())

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
		input:  newInputReader(os.Stdin),
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

// WatchResize listens for SIGWINCH and updates the terminal dimensions.
// It runs in a background goroutine and stops when ctx is cancelled.
// Call once from the editor's Run() to keep Width()/Height() accurate on resize.
func (t *Terminal) WatchResize() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			_ = t.UpdateSize()
		}
	}()
}

// --- Screen buffer methods (build frame in memory, then flush) ---

// appends a string to the in-memory screen buffer.
func (t *Terminal) WriteStr(s string) {
	t.builder.WriteString(s)
}

// resets the in-memory screen buffer.
func (t *Terminal) ClearBuffer() {
	t.builder.Reset()
}

// returns the accumulated screen buffer contents.
func (t *Terminal) BufferString() string {
	return t.builder.String()
}

// writes hide-cursor and move-home to the buffer (start of frame).
func (t *Terminal) PrepareScreen() {
	t.WriteStr(ansi.HideCursor)
	t.WriteStr(ansi.MoveCursorHome)
}

// writes cursor-position and show-cursor to the buffer (end of frame).
func (t *Terminal) FinalizeScreen(row, col int) {
	t.WriteStr(ansi.MoveCursorTo(row, col))
	t.WriteStr(ansi.ShowCursor)
}

// writes a cursor-movement escape to the buffer.
func (t *Terminal) MoveCursorTo(row, col int) {
	t.WriteStr(ansi.MoveCursorTo(row, col))
}

// writes a clear-to-end-of-line escape to the buffer.
func (t *Terminal) ClearToLineEnd() {
	t.WriteStr(ansi.ClearToLineEnd)
}

// writes a format-reset escape to the buffer.
func (t *Terminal) ResetFormat() {
	t.WriteStr(ansi.ResetFormat)
}
