package terminal

import (
	"strings"

	"github.com/AdityaKrSingh26/Glime/pkg/ansi"
)

// Build the entire frame in memory first
// Then write it to the terminal in one single operation

// represents an in-memory buffer for building screen content before writing it to the terminal.
// This reduces flicker by doing a single write operation instead of many small writes.
type ScreenBuffer struct {
	builder strings.Builder
}

func NewScreenBuffer() *ScreenBuffer {
	return &ScreenBuffer{
		builder: strings.Builder{},
	}
}

// sb is a receiver; allows function to be called by a dot method
func (sb *ScreenBuffer) Write(s string) {
	sb.builder.WriteString(s)
}

func (sb *ScreenBuffer) Clear() {
	sb.builder.Reset()
}

func (sb *ScreenBuffer) String() string {
	return sb.builder.String()
}

// add a common initialization sequence
// - Hide cursor
// - Move to home position
// this should be called at the start of building a frame
func (sb *ScreenBuffer) PrepareScreen() {
	sb.Write(ansi.HideCursor)
	sb.Write(ansi.MoveCursorHome)
}

// add a common finalization sequence
// - move cursor to a specific position
// - show cursor
// this should be called at end of building frame
func (sb *ScreenBuffer) FinalizeScreen(row, col int) {
	sb.Write(ansi.MoveCursorTo(row, col))
	sb.Write(ansi.ShowCursor)
}

// adds a cursor movement command to the buffer.
func (sb *ScreenBuffer) MoveCursor(row, col int) {
	sb.Write(ansi.MoveCursorTo(row, col))
}

// adds a clear-to-line-end command to the buffer.
func (sb *ScreenBuffer) ClearToLineEnd() {
	sb.Write(ansi.ClearToLineEnd)
}

// Reset formatting to default.
func (sb *ScreenBuffer) ResetFormat() {
	sb.Write(ansi.ResetFormat)
}
