// provides ANSI escape sequence constants and utilities for terminal control operations.
package ansi

import "fmt"

// FYI : An ANSI escape sequence looks like this:
// " ESC [ parameters command "
// \xlb -> ESC

// Screen control sequences (ANSI escape codes)
const (
	ClearScreen            = "\x1b[2J"     // Clear entire screen
	ClearLine              = "\x1b[2K"     // Clear current line
	ClearToLineEnd         = "\x1b[K"      // Clear from cursor to line end
	ClearToScreenEnd       = "\x1b[J"      // Clear from cursor to screen end
	MoveCursorHome         = "\x1b[H"      // Move cursor to (1,1)
	HideCursor             = "\x1b[?25l"   // Hide cursor
	ShowCursor             = "\x1b[?25h"   // Show cursor
	SaveCursor             = "\x1b[s"      // Save cursor position
	RestoreCursor          = "\x1b[u"      // Restore cursor position
	EnableAlternateBuffer  = "\x1b[?1049h" // Switch to alternate screen buffer
	DisableAlternateBuffer = "\x1b[?1049l" // Return to main screen buffer
)

// Text formatting sequences (ANSI escape codes)
const (
	ResetFormat = "\x1b[0m" // Reset all text formatting
	Bold        = "\x1b[1m" // Enable bold text
	Dim         = "\x1b[2m" // Enable dim/faint text
	Italic      = "\x1b[3m" // Enable italic text
	Underline   = "\x1b[4m" // Enable underlined text
	Inverse     = "\x1b[7m" // Invert foreground and background colors
)

// Standard color codes
const (
	// Foreground colors
	FgBlack   = "\x1b[30m"
	FgRed     = "\x1b[31m"
	FgGreen   = "\x1b[32m"
	FgYellow  = "\x1b[33m"
	FgBlue    = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan    = "\x1b[36m"
	FgWhite   = "\x1b[37m"

	// Background colors
	BgBlack   = "\x1b[40m"
	BgRed     = "\x1b[41m"
	BgGreen   = "\x1b[42m"
	BgYellow  = "\x1b[43m"
	BgBlue    = "\x1b[44m"
	BgMagenta = "\x1b[45m"
	BgCyan    = "\x1b[46m"
	BgWhite   = "\x1b[47m"
)

// MoveCursorTo moves cursor to specific row and column (1-indexed).
// fmt.Print(ansi.MoveCursorTo(5, 10))  // Move to row 5, column 10
func MoveCursorTo(row, col int) string {
	return fmt.Sprintf("\x1b[%d;%dH", row, col)
}

// moves cursor up by n lines.
func MoveCursorUp(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dA", n)
}

// moves cursor down by n lines.
func MoveCursorDown(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dB", n)
}

// moves cursor right by n columns.
func MoveCursorRight(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dC", n)
}

// moves cursor left by n columns.
func MoveCursorLeft(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dD", n)
}

// sets foreground color using 8-bit color (0-255).
func SetFgColor(color int) string {
	return fmt.Sprintf("\x1b[38;5;%dm", color)
}

// sets background color using 8-bit color (0-255).
func SetBgColor(color int) string {
	return fmt.Sprintf("\x1b[48;5;%dm", color)
}

// sets foreground color using RGB values (0-255).
func SetFgRGB(r, g, b int) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

// sets background color using RGB values (0-255).
func SetBgRGB(r, g, b int) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// requests cursor position (returns escape sequence that terminal responds to).
func GetCursorPosition() string {
	return "\x1b[6n"
}
