// Package ui provides user interface rendering components.
package ui

import (
	"fmt"
	"strings"

	"github.com/AdityaKrSingh26/Glime/internal/syntax"
	"github.com/AdityaKrSingh26/Glime/internal/terminal"
	"github.com/AdityaKrSingh26/Glime/pkg/ansi"
)

// responsible for rendering the editor UI to the terminal.
type Renderer struct {
	terminal    *terminal.Terminal
	buffer      *terminal.ScreenBuffer
	theme       Theme
	highlighter *syntax.Highlighter
}

func NewRenderer(term *terminal.Terminal) *Renderer {
	return &Renderer{
		terminal: term,
		buffer:   terminal.NewScreenBuffer(),
		theme:    DefaultTheme(),
	}
}

// sets the language for syntax highlighting based on file path.
func (r *Renderer) SetLanguage(filePath string) {
	lang := syntax.DetectLanguage(filePath)
	if lang != nil {
		colorTheme := syntax.ColorTheme{
			Keyword:  r.theme.Keyword,
			String:   r.theme.String,
			Comment:  r.theme.Comment,
			Number:   r.theme.Number,
			Function: r.theme.Function,
			Type:     r.theme.Type,
			Operator: r.theme.Operator,
			Builtin:  r.theme.Builtin,
		}
		r.highlighter = syntax.NewHighlighter(lang, colorTheme)
	} else {
		r.highlighter = nil
	}
}

// represents a highlighted range on a line.
type MatchRange struct {
	ColStart int
	ColEnd   int
}

// provides the data needed to render the editor.
type EditorView struct {
	Buffer     BufferView
	Cursor     CursorView
	Mode       ModeView
	Message    string
	TermWidth  int
	TermHeight int
	TotalLines int // total lines in buffer for gutter width calculation

	// Search highlighting
	SearchMatches map[int][]MatchRange // row -> list of match ranges
	SearchActive  bool

	// Bracket matching
	BracketMatch *BracketMatchView // nil if no match
}

// holds the position of a matching bracket for rendering.
type BracketMatchView struct {
	Row int
	Col int
}

// provides buffer data for rendering.
type BufferView struct {
	Lines      []string
	FileName   string
	IsModified bool
}

// provides cursor position data for rendering.
type CursorView struct {
	Row       int
	Col       int
	RowOffset int
	ColOffset int
}

// provides mode information for rendering.
type ModeView struct {
	Name string
}

// renders the entire editor screen using the provided view data.
func (r *Renderer) Render(view EditorView) error {

	r.buffer.Clear()
	r.buffer.PrepareScreen()
	r.renderBuffer(view)
	r.renderStatusBar(view)
	r.renderMessageBar(view)

	// Calculate cursor screen position (1-indexed for terminal)
	// consider the gutter width
	gutterWidth := GutterWidth(view.TotalLines)
	screenRow := (view.Cursor.Row - view.Cursor.RowOffset) + 1
	screenCol := (view.Cursor.Col - view.Cursor.ColOffset) + 1 + gutterWidth

	// Finalize screen (position cursor, show cursor)
	r.buffer.FinalizeScreen(screenRow, screenCol)

	// Write entire buffer to terminal in one operation
	return r.terminal.Write(r.buffer.String())
}

// calculates the width needed for line numbers based on total lines.
func GutterWidth(totalLines int) int {
	if totalLines < 10 {
		return 4
	} else if totalLines < 100 {
		return 5
	} else if totalLines < 1000 {
		return 6
	} else if totalLines < 10000 {
		return 7
	}
	return 8
}

// renders a line number in the gutter with proper formatting.
func (r *Renderer) renderLineNumber(lineNum, width int, isCurrent bool) {
	// Set colors based on whether this is the current line
	if isCurrent {
		r.buffer.Write(ansi.SetFgColor(r.theme.CurrentLineNumber))
		r.buffer.Write(ansi.Bold)
	} else {
		r.buffer.Write(ansi.SetFgColor(r.theme.LineNumber))
	}

	// Format line number right-aligned with padding
	lineStr := fmt.Sprintf("%*d ", width-1, lineNum)
	r.buffer.Write(lineStr)

	// Reset formatting
	r.buffer.ResetFormat()
}

// renders an empty line indicator (tilde) with gutter.
func (r *Renderer) renderEmptyLine(gutterWidth int) {
	// Render empty gutter space
	r.buffer.Write(strings.Repeat(" ", gutterWidth))

	// Render tilde
	r.buffer.Write(ansi.SetFgColor(r.theme.EmptyLine))
	r.buffer.Write("~")
	r.buffer.ResetFormat()
}

// renders the visible portion of the text buffer with line numbers.
func (r *Renderer) renderBuffer(view EditorView) {
	// Calculate visible area (excluding status and message bars)
	visibleRows := view.TermHeight - 2

	// Calculate gutter width based on total lines
	gutterWidth := GutterWidth(view.TotalLines)

	// Calculate available width for text after gutter
	textWidth := view.TermWidth - gutterWidth

	for y := 0; y < visibleRows; y++ {
		fileRow := y + view.Cursor.RowOffset

		// Move cursor to this row
		r.buffer.MoveCursor(y+1, 1)

		if fileRow >= len(view.Buffer.Lines) {
			// Past end of file - show empty gutter and tilde
			r.renderEmptyLine(gutterWidth)
		} else {
			// Render line number in gutter
			lineNum := fileRow + 1
			isCurrent := fileRow == view.Cursor.Row
			r.renderLineNumber(lineNum, gutterWidth, isCurrent)

			// Render line content with syntax highlighting and horizontal scrolling
			line := view.Buffer.Lines[fileRow]

			// Apply syntax highlighting to the full line, then extract visible portion
			var displayLine string
			if r.highlighter != nil {
				highlighted := r.highlighter.Highlight(line)
				displayLine = extractVisiblePortion(highlighted, view.Cursor.ColOffset, textWidth)
			} else if view.Cursor.ColOffset < len(line) {
				visibleEnd := view.Cursor.ColOffset + textWidth
				if visibleEnd > len(line) {
					visibleEnd = len(line)
				}
				displayLine = line[view.Cursor.ColOffset:visibleEnd]
			}

			// Apply search highlighting on top
			if view.SearchActive {
				if matches, ok := view.SearchMatches[fileRow]; ok && len(matches) > 0 {
					displayLine = r.applySearchHighlight(displayLine, matches, view.Cursor.ColOffset, textWidth)
				}
			}

			// Apply bracket match highlighting
			if view.BracketMatch != nil && view.BracketMatch.Row == fileRow {
				matchCol := view.BracketMatch.Col - view.Cursor.ColOffset
				if matchCol >= 0 && matchCol < textWidth {
					displayLine = r.applyBracketHighlight(displayLine, matchCol)
				}
			}

			r.buffer.Write(displayLine)
		}

		// Clear to end of line (in case line got shorter)
		r.buffer.ClearToLineEnd()
	}
}

// applies yellow background to search matches within the visible portion.
func (r *Renderer) applySearchHighlight(
	displayLine string,
	matches []MatchRange,
	colOffset,
	textWidth int,
) string {

	// a map of visible positions that should be highlighted
	highlightPositions := make(map[int]bool)
	for _, m := range matches {
		for col := m.ColStart; col < m.ColEnd; col++ {
			visCol := col - colOffset
			if visCol >= 0 && visCol < textWidth {
				highlightPositions[visCol] = true
			}
		}
	}

	if len(highlightPositions) == 0 {
		return displayLine
	}

	// Walk through displayLine, tracking visible character position
	var result strings.Builder
	visPos := 0
	inEscape := false
	highlighted := false

	for i := 0; i < len(displayLine); i++ {
		ch := displayLine[i]

		if ch == '\x1b' {
			// Start of ANSI escape sequence
			if highlighted {
				result.WriteString(ansi.ResetFormat)
				highlighted = false
			}
			inEscape = true
			result.WriteByte(ch)
			continue
		}

		if inEscape {
			result.WriteByte(ch)
			if ch == 'm' {
				inEscape = false
			}
			continue
		}

		// Visible character
		if highlightPositions[visPos] {
			if !highlighted {
				result.WriteString(ansi.SetBgColor(r.theme.SearchHighlight))
				result.WriteString(ansi.SetFgColor(0)) // black text on yellow
				highlighted = true
			}
		} else if highlighted {
			result.WriteString(ansi.ResetFormat)
			highlighted = false
		}

		result.WriteByte(ch)
		visPos++
	}

	if highlighted {
		result.WriteString(ansi.ResetFormat)
	}

	return result.String()
}

// highlights a single character at the given visible column.
func (r *Renderer) applyBracketHighlight(displayLine string, visCol int) string {
	var result strings.Builder
	curVisPos := 0
	inEscape := false

	for i := 0; i < len(displayLine); i++ {
		ch := displayLine[i]

		if ch == '\x1b' {
			inEscape = true
			result.WriteByte(ch)
			continue
		}

		if inEscape {
			result.WriteByte(ch)
			if ch == 'm' {
				inEscape = false
			}
			continue
		}

		if curVisPos == visCol {
			result.WriteString(ansi.SetBgColor(r.theme.BracketMatch))
			result.WriteByte(ch)
			result.WriteString(ansi.ResetFormat)
		} else {
			result.WriteByte(ch)
		}
		curVisPos++
	}

	return result.String()
}

// extracts the visible portion of an ANSI-highlighted string.
func extractVisiblePortion(highlighted string, colOffset, width int) string {
	if colOffset == 0 && width >= len(highlighted) {
		// Fast path: no scrolling and line fits
		return highlighted
	}

	var result strings.Builder
	visPos := 0
	visWritten := 0
	inEscape := false
	// Track the last ANSI code seen before the visible region for color continuity
	var activeColor strings.Builder

	for i := 0; i < len(highlighted); i++ {
		ch := highlighted[i]

		if ch == '\x1b' {
			inEscape = true
			if visPos < colOffset {
				activeColor.Reset()
				activeColor.WriteByte(ch)
			} else if visWritten < width {
				result.WriteByte(ch)
			}
			continue
		}

		if inEscape {
			if visPos < colOffset {
				activeColor.WriteByte(ch)
			} else if visWritten < width {
				result.WriteByte(ch)
			}
			if ch == 'm' {
				inEscape = false
			}
			continue
		}

		// Visible character
		if visPos >= colOffset && visWritten < width {
			if visWritten == 0 && activeColor.Len() > 0 {
				// Prepend the active color at the start of visible region
				result.WriteString(activeColor.String())
			}
			result.WriteByte(ch)
			visWritten++
		}
		visPos++

		if visWritten >= width {
			break
		}
	}

	return result.String()
}

// renders the status bar at the bottom of the screen.
func (r *Renderer) renderStatusBar(view EditorView) {
	// Position at status bar row
	statusRow := view.TermHeight - 1
	r.buffer.MoveCursor(statusRow, 1)

	// Calculate percentage
	percentage := calculatePercentage(view.Cursor.Row, len(view.Buffer.Lines))

	// Create enhanced status bar
	statusLine := EnhancedStatusBar(
		r.theme,
		view.Mode.Name,
		view.Buffer.FileName,
		view.Buffer.IsModified,
		view.Cursor.Row+1,
		view.Cursor.Col+1,
		percentage,
		view.TermWidth,
	)

	// Write status bar
	r.buffer.Write(statusLine)

	// Clear to end of line (in case terminal is wider)
	r.buffer.ClearToLineEnd()
}

// renders the message/command bar at the very bottom.
func (r *Renderer) renderMessageBar(view EditorView) {
	// Position at message bar row
	messageRow := view.TermHeight
	r.buffer.MoveCursor(messageRow, 1)

	message := view.Message

	// Truncate message if too long
	if len(message) > view.TermWidth {
		message = message[:view.TermWidth]
	}

	r.buffer.Write(message)
	r.buffer.ClearToLineEnd()
}

// calculates the percentage through the file.
func calculatePercentage(currentRow, totalRows int) int {
	if totalRows == 0 {
		return 100
	}

	percentage := (currentRow * 100) / totalRows

	// Clamp to 0-100
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	return percentage
}
