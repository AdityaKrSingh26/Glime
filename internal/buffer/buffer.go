package buffer

import (
	"fmt"
	"strings"
)

// this is to provide text buffer management to text editor
// represent in memory content of file being edited

// Buffer represents the text content being edited.
// It stores text as a slice of lines and tracks the modification state.
// Note : not safe to concurrent access
type Buffer struct {
	lines    []string
	modified bool   // Whether buffer has unsaved changes
	filePath string // Associated file path (empty for new buffers)
}

func New() *Buffer {
	return &Buffer{
		lines:    []string{""},
		modified: false,
		filePath: "",
	}
}

// creates buffer from existing lines
func NewFromLines(lines []string, filePath string) *Buffer {
	if len(lines) == 0 {
		lines = []string{""}
	}

	return &Buffer{
		lines:    lines,
		modified: false,
		filePath: filePath,
	}
}

// return the number of lines
func (b *Buffer) NumLines() int {
	return len(b.lines)
}

// returns the content of a specific line that is inbound
func (b *Buffer) GetLine(row int) (string, error) {
	if row < 0 || row >= len(b.lines) {
		return "", fmt.Errorf("line %d out of bounds (0-%d)", row, len(b.lines)-1)
	}
	return b.lines[row], nil
}

// returns all lines in the buffer.
func (b *Buffer) GetLines() []string {
	return b.lines
}

// returns the length of the specified line.
func (b *Buffer) LineLength(row int) (int, error) {
	if row < 0 || row >= len(b.lines) {
		return 0, fmt.Errorf("line %d out of bounds", row)
	}
	return len(b.lines[row]), nil
}

// insert a new character at a give row, col
// return error if row/col is out of bounds
func (b *Buffer) InsertChar(row, col int, ch rune) error {
	if row < 0 || row >= len(b.lines) {
		return fmt.Errorf("row %d out of bounds", row)
	}

	line := b.lines[row]
	if col < 0 || col > len(line) {
		return fmt.Errorf("col %d out of bounds (0-%d)", col, len(line))
	}

	// Insert character at position
	b.lines[row] = line[:col] + string(ch) + line[col:]
	b.modified = true
	return nil
}

// delete a char at a give row, col
// return nil if end of line || error if row/char out of bound
func (b *Buffer) DeleteChar(row, col int) error {
	if row < 0 || row >= len(b.lines) {
		return fmt.Errorf("row %d out of bounds", row)
	}

	line := b.lines[row]
	if col < 0 || col > len(line) {
		return fmt.Errorf("col %d out of bounds", col)
	}

	// end of line, nothing to delete
	if col == len(line) {
		return nil
	}

	// Delete character at position
	b.lines[row] = line[:col] + line[col+1:]
	b.modified = true
	return nil
}

// inserts a new empty line at the specified row.
// existing lines are shifted down.
func (b *Buffer) InsertLine(row int) error {
	if row < 0 || row > len(b.lines) {
		return fmt.Errorf("row %d out of bounds (0-%d)", row, len(b.lines))
	}

	// Insert empty line
	b.lines = append(b.lines[:row], append([]string{""}, b.lines[row:]...)...)
	b.modified = true
	return nil
}

// deletes the specified line
func (b *Buffer) DeleteLine(row int) error {
	if row < 0 || row >= len(b.lines) {
		return fmt.Errorf("row %d out of bounds", row)
	}

	// If only one line, make it empty instead of deleting
	if len(b.lines) == 1 {
		b.lines[0] = ""
		b.modified = true
		return nil
	}

	// Delete line
	b.lines = append(b.lines[:row], b.lines[row+1:]...)
	b.modified = true
	return nil
}

// splits the line at the specified position.
// text after col becomes a new line inserted below.
// used when pressing Enter in the middle of a line.
func (b *Buffer) SplitLine(row, col int) error {
	if row < 0 || row >= len(b.lines) {
		return fmt.Errorf("row %d out of bounds", row)
	}

	line := b.lines[row]
	if col < 0 || col > len(line) {
		return fmt.Errorf("col %d out of bounds", col)
	}

	// Split line at column
	before := line[:col]
	after := line[col:]

	b.lines[row] = before
	b.lines = append(b.lines[:row+1], append([]string{after}, b.lines[row+1:]...)...)
	b.modified = true
	return nil
}

// joins the specified line with the next line
func (b *Buffer) JoinLines(row int) error {
	if row < 0 || row >= len(b.lines)-1 {
		return fmt.Errorf("cannot join line %d", row)
	}

	// Join with next line
	b.lines[row] = b.lines[row] + b.lines[row+1]
	b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
	b.modified = true
	return nil
}

// Backspace deletes the character before the cursor.
// start of a line (col=0), it joins with the previous line.
func (b *Buffer) Backspace(row, col int) (newRow, newCol int, err error) {
	if row < 0 || row >= len(b.lines) {
		return row, col, fmt.Errorf("row %d out of bounds", row)
	}

	// If at start of first line, nothing to delete
	if row == 0 && col == 0 {
		return row, col, nil
	}

	// If at start of line, join with previous line
	if col == 0 {
		prevLineLen := len(b.lines[row-1])
		if err := b.JoinLines(row - 1); err != nil {
			return row, col, err
		}
		return row - 1, prevLineLen, nil
	}

	// Delete previous character
	line := b.lines[row]
	if col > len(line) {
		col = len(line)
	}

	b.lines[row] = line[:col-1] + line[col:]
	b.modified = true
	return row, col - 1, nil
}

// replaces the content of the specified line.
func (b *Buffer) SetLine(row int, text string) error {
	if row < 0 || row >= len(b.lines) {
		return fmt.Errorf("row %d out of bounds", row)
	}
	b.lines[row] = text
	b.modified = true
	return nil
}

// inserts a line with the given text at the specified row.
// existing lines are shifted down.
func (b *Buffer) InsertLineWithContent(row int, text string) error {
	if row < 0 || row > len(b.lines) {
		return fmt.Errorf("row %d out of bounds (0-%d)", row, len(b.lines))
	}
	b.lines = append(b.lines[:row], append([]string{text}, b.lines[row:]...)...)
	b.modified = true
	return nil
}

func (b *Buffer) IsModified() bool {
	return b.modified
}

func (b *Buffer) SetModified(modified bool) {
	b.modified = modified
}

func (b *Buffer) FilePath() string {
	return b.filePath
}

func (b *Buffer) SetFilePath(path string) {
	b.filePath = path
}

// returns just the filename (without directory path).
func (b *Buffer) FileName() string {
	if b.filePath == "" {
		return "[No Name]"
	}

	parts := strings.Split(b.filePath, "/")
	return parts[len(parts)-1]
}

func (b *Buffer) IsEmpty() bool {
	return len(b.lines) == 1 && b.lines[0] == ""
}

func (b *Buffer) String() string {
	return strings.Join(b.lines, "\n")
}
