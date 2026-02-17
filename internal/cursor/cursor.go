package cursor

import "github.com/AdityaKrSingh26/Glime/internal/buffer"

// represents the cursor position and viewport state
// tracks both absolute position and offset in buffer
type Cursor struct {
	row        int
	col        int
	desiredCol int // Desired column when moving up/down
	rowOffset  int // Vertical scroll offset (first visible row)
	colOffset  int // Horizontal scroll offset (first visible column)
}

// create a new cursor at (0,0)
func New() *Cursor {
	return &Cursor{
		row:        0,
		col:        0,
		desiredCol: 0,
		rowOffset:  0,
		colOffset:  0,
	}
}

func (c *Cursor) Row() int {
	return c.row
}

func (c *Cursor) Col() int {
	return c.col
}

func (c *Cursor) RowOffset() int {
	return c.rowOffset
}

func (c *Cursor) ColOffset() int {
	return c.colOffset
}

// sets the cursor position directly; does not check bounds
func (c *Cursor) SetPosition(row, col int) {
	c.row = row
	c.col = col
	c.desiredCol = col
}

// moves cursor to a specified position and also do a bounding check
func (c *Cursor) MoveTo(row, col int, buff *buffer.Buffer) {
	if row < 0 {
		row = 0
	}
	if row >= buff.NumLines() {
		row = buff.NumLines() - 1
	}

	lineLen, _ := buff.LineLength(row)
	if col < 0 {
		col = 0
	}
	if col > lineLen {
		col = lineLen
	}

	c.row = row
	c.col = col
	c.desiredCol = col
}

// adjust the column to be in lines bound
// maintain desired column while moving up and down
func (c *Cursor) clampColumn(buff *buffer.Buffer) {
	lineLen, _ := buff.LineLength()

	if c.desiredCol <= lineLen {
		c.col = c.desiredCol
	} else {
		c.col = lineLen
	}
}

// moves the cursor up by one line.
func (c *Cursor) MoveUp(buf *buffer.Buffer) {
	if c.row > 0 {
		c.row--
		c.clampColumn(buf)
	}
}

// moves the cursor down by one line.
func (c *Cursor) MoveDown(buf *buffer.Buffer) {
	if c.row < buf.NumLines()-1 {
		c.row++
		c.clampColumn(buf)
	}
}

// moves the cursor left by one column
// if cursor is at start of a line, move to the end of previous line
func (c *Cursor) MoveLeft(buf *buffer.Buffer) {
	if c.col > 0 {
		c.col--
		c.desiredCol = c.col
	} else if c.row > 0 {
		// Move to end of previous line
		c.row--
		lineLen, _ := buf.LineLength(c.row)
		c.col = lineLen
		c.desiredCol = c.col
	}
}

// moves the cursor to right by one column
// if cursor is at the end of a line, move it to start of next line
func (c *Cursor) MoveRight(buf *buffer.Buffer) {
	lineLen, _ := buf.LineLength(c.row)

	if c.col < lineLen {
		c.col++
		c.desiredCol = c.col
	} else if c.row < buf.NumLines()-1 {
		// Move to start of next line
		c.row++
		c.col = 0
		c.desiredCol = 0
	}
}

// moves the cursor to the beginning of the current line.
func (c *Cursor) MoveToLineStart() {
	c.col = 0
	c.desiredCol = 0
}

// moves the cursor to the end of the current line.
func (c *Cursor) MoveToLineEnd(buf *buffer.Buffer) {
	lineLen, _ := buf.LineLength(c.row)
	c.col = lineLen
	c.desiredCol = lineLen
}

// moves the cursor to the first line of the buffer.
func (c *Cursor) MoveToFirstLine() {
	c.row = 0
	c.col = 0
	c.desiredCol = 0
}

// moves the cursor to the last line of the buffer.
func (c *Cursor) MoveToLastLine(buf *buffer.Buffer) {
	c.row = buf.NumLines() - 1
	c.col = 0
	c.desiredCol = 0
}

// moves the cursor up by one page (terminal height).
func (c *Cursor) PageUp(buf *buffer.Buffer, pageSize int) {
	c.row -= pageSize
	if c.row < 0 {
		c.row = 0
	}
	c.clampColumn(buf)
}

// moves the cursor down by one page (terminal height).
func (c *Cursor) PageDown(buf *buffer.Buffer, pageSize int) {
	c.row += pageSize
	if c.row >= buf.NumLines() {
		c.row = buf.NumLines() - 1
	}
	c.clampColumn(buf)
}

// update the scroll offset to ensure the cursor is visible
//   - screenRows: number of visible rows (terminal height minus status bars)
//   - screenCols: number of visible columns (terminal width)
func (c *Cursor) UpdateScroll(screenRows, screenCols int) {
	// vertical scrolling
	if c.row < c.rowOffset {
		c.rowOffset = c.row
	}
	if c.row >= c.rowOffset+screenRows {
		c.rowOffset = c.row - screenRows + 1
	}

	// Horizontal scrolling
	if c.col < c.colOffset {
		c.colOffset = c.col
	}
	if c.col >= c.colOffset+screenCols {
		c.colOffset = c.col - screenCols + 1
	}
}
