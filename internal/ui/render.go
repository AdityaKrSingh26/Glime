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
	theme       Theme
	highlighter *syntax.Highlighter
}

func NewRenderer(term *terminal.Terminal) *Renderer {
	return &Renderer{
		terminal: term,
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

// represents a single entry in the explorer listing.
type ExplorerViewEntry struct {
	DisplayName string
	IsDir       bool
	Size        int64
}

// holds the data needed to render the file explorer.
type ExplorerView struct {
	Dir       string
	Entries   []ExplorerViewEntry
	CursorRow int
	RowOffset int
}

// provides the data needed to render the editor.
type EditorView struct {
	Lines      []string
	FileName   string
	IsModified bool
	CursorRow  int
	CursorCol  int
	RowOffset  int
	ColOffset  int
	ModeName   string
	Message    string
	TermWidth  int
	TermHeight int
	TotalLines int

	// Search highlighting
	SearchMatches map[int][]MatchRange // row -> list of match ranges
	SearchActive  bool

	// Bracket matching
	BracketMatch *BracketMatchView // nil if no match

	// Explorer mode
	IsExplorer bool
	Explorer   ExplorerView
}

// holds the position of a matching bracket for rendering.
type BracketMatchView struct {
	Row int
	Col int
}

// renders the entire editor screen using the provided view data.
func (r *Renderer) Render(view EditorView) error {

	r.terminal.ClearBuffer()
	r.terminal.PrepareScreen()

	var screenRow, screenCol int

	if view.IsExplorer {
		r.renderExplorer(view)
		// Explorer cursor: 2 header rows + offset within visible entries
		headerRows := 2
		screenRow = view.Explorer.CursorRow - view.Explorer.RowOffset + headerRows + 1
		screenCol = 1
	} else {
		r.renderBuffer(view)
		// Calculate cursor screen position (1-indexed for terminal)
		// consider the gutter width
		gutterWidth := GutterWidth(view.TotalLines)
		screenRow = (view.CursorRow - view.RowOffset) + 1
		screenCol = (view.CursorCol - view.ColOffset) + 1 + gutterWidth
	}

	r.renderStatusBar(view)
	r.renderMessageBar(view)

	// Finalize screen (position cursor, show cursor)
	r.terminal.FinalizeScreen(screenRow, screenCol)

	// Write entire buffer to terminal in one operation
	return r.terminal.Write(r.terminal.BufferString())
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
		r.terminal.WriteStr(ansi.SetFgColor(r.theme.CurrentLineNumber))
		r.terminal.WriteStr(ansi.Bold)
	} else {
		r.terminal.WriteStr(ansi.SetFgColor(r.theme.LineNumber))
	}

	// Format line number right-aligned with padding
	lineStr := fmt.Sprintf("%*d ", width-1, lineNum)
	r.terminal.WriteStr(lineStr)

	// Reset formatting
	r.terminal.ResetFormat()
}

// renders an empty line indicator (tilde) with gutter.
func (r *Renderer) renderEmptyLine(gutterWidth int) {
	// Render empty gutter space
	r.terminal.WriteStr(strings.Repeat(" ", gutterWidth))

	// Render tilde
	r.terminal.WriteStr(ansi.SetFgColor(r.theme.EmptyLine))
	r.terminal.WriteStr("~")
	r.terminal.ResetFormat()
}

// renders the file explorer view (netrw-style).
func (r *Renderer) renderExplorer(view EditorView) {
	ev := view.Explorer
	listingRows := view.TermHeight - 4 // 2 header + 1 status + 1 message

	// Row 1: header
	r.renderExplorerHeader(ev.Dir, view.TermWidth)

	// Row 2: separator
	r.terminal.MoveCursorTo(2, 1)
	r.terminal.WriteStr(ansi.SetFgColor(r.theme.Border))
	r.terminal.WriteStr(strings.Repeat("─", view.TermWidth))
	r.terminal.ResetFormat()

	// Rows 3+: entries
	for y := 0; y < listingRows; y++ {
		r.terminal.MoveCursorTo(y+3, 1)

		idx := y + ev.RowOffset
		if idx >= len(ev.Entries) {
			r.renderEmptyLine(0)
			continue
		}

		r.renderExplorerEntry(ev.Entries[idx], idx == ev.CursorRow, view.TermWidth)
	}
}

func (r *Renderer) renderExplorerHeader(dir string, termWidth int) {
	r.terminal.MoveCursorTo(1, 1)
	r.terminal.WriteStr(ansi.SetFgColor(r.theme.Function))
	r.terminal.WriteStr(ansi.Bold)
	r.terminal.WriteStr(" netrw")
	r.terminal.ResetFormat()
	r.terminal.WriteStr("  ")

	// truncate long paths from the left
	if len(dir) > termWidth-10 {
		dir = "..." + dir[len(dir)-termWidth+13:]
	}
	r.terminal.WriteStr(ansi.SetFgColor(r.theme.Comment))
	r.terminal.WriteStr(dir)
	r.terminal.ResetFormat()
	r.terminal.ClearToLineEnd()
}

func (r *Renderer) renderExplorerEntry(entry ExplorerViewEntry, isSelected bool, termWidth int) {
	// pick prefix and color based on dir vs file
	prefix := "   "
	color := r.theme.String
	if entry.IsDir {
		prefix = " > "
		color = r.theme.Function
	}

	// selected entries get inverse highlight
	if isSelected {
		r.terminal.WriteStr(ansi.SetBgColor(r.theme.StatusModeBg))
		r.terminal.WriteStr(ansi.SetFgColor(r.theme.StatusFg))
	} else {
		r.terminal.WriteStr(ansi.SetFgColor(color))
	}
	if entry.IsDir {
		r.terminal.WriteStr(ansi.Bold)
	}

	line := prefix + entry.DisplayName
	if len(line) > termWidth {
		line = line[:termWidth]
	}
	r.terminal.WriteStr(line)

	r.terminal.ResetFormat()
	r.terminal.ClearToLineEnd()
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
		fileRow := y + view.RowOffset

		// Move cursor to this row
		r.terminal.MoveCursorTo(y+1, 1)

		if fileRow >= len(view.Lines) {
			// Past end of file - show empty gutter and tilde
			r.renderEmptyLine(gutterWidth)
		} else {
			// Render line number in gutter
			lineNum := fileRow + 1
			isCurrent := fileRow == view.CursorRow
			r.renderLineNumber(lineNum, gutterWidth, isCurrent)

			// Render line content with syntax highlighting and horizontal scrolling
			line := view.Lines[fileRow]

			// Apply syntax highlighting to the full line, then extract visible portion
			var displayLine string
			if r.highlighter != nil {
				highlighted := r.highlighter.Highlight(line)
				displayLine = extractVisiblePortion(highlighted, view.ColOffset, textWidth)
			} else {
				runes := []rune(line)
				if view.ColOffset < len(runes) {
					visibleEnd := view.ColOffset + textWidth
					if visibleEnd > len(runes) {
						visibleEnd = len(runes)
					}
					displayLine = string(runes[view.ColOffset:visibleEnd])
				}
			}

			// Apply search highlighting on top
			if view.SearchActive {
				if matches, ok := view.SearchMatches[fileRow]; ok && len(matches) > 0 {
					displayLine = r.applySearchHighlight(displayLine, matches, view.ColOffset, textWidth)
				}
			}

			// Apply bracket match highlighting
			if view.BracketMatch != nil && view.BracketMatch.Row == fileRow {
				matchCol := view.BracketMatch.Col - view.ColOffset
				if matchCol >= 0 && matchCol < textWidth {
					displayLine = r.applyBracketHighlight(displayLine, matchCol)
				}
			}

			r.terminal.WriteStr(displayLine)
		}

		// Clear to end of line (in case line got shorter)
		r.terminal.ClearToLineEnd()
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
	r.terminal.MoveCursorTo(statusRow, 1)

	// Calculate percentage
	percentage := calculatePercentage(view.CursorRow, len(view.Lines))

	// Create enhanced status bar
	statusLine := EnhancedStatusBar(
		r.theme,
		view.ModeName,
		view.FileName,
		view.IsModified,
		view.CursorRow+1,
		view.CursorCol+1,
		percentage,
		view.TermWidth,
	)

	// Write status bar
	r.terminal.WriteStr(statusLine)

	// Clear to end of line (in case terminal is wider)
	r.terminal.ClearToLineEnd()
}

// renders the message/command bar at the very bottom.
func (r *Renderer) renderMessageBar(view EditorView) {
	// Position at message bar row
	messageRow := view.TermHeight
	r.terminal.MoveCursorTo(messageRow, 1)

	message := view.Message

	// Truncate message if too long
	if len(message) > view.TermWidth {
		message = message[:view.TermWidth]
	}

	r.terminal.WriteStr(message)
	r.terminal.ClearToLineEnd()
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
