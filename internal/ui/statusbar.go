package ui

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AdityaKrSingh26/Glime/internal/syntax"
	"github.com/AdityaKrSingh26/Glime/pkg/ansi"
)

// matches any ANSI escape sequence.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// returns the number of visible (non-ANSI) characters in s.
func visibleLen(s string) int {
	return len(ansiEscape.ReplaceAllString(s, ""))
}

// creates a colorful, segmented status bar with modern styling.
func EnhancedStatusBar(
	theme Theme,
	mode,
	fileName string,
	modified bool,
	row,
	col,
	percentage,
	width int,
) string {

	var result strings.Builder

	icon := getModeIcon(mode)
	modeText := fmt.Sprintf(" %s %s ", icon, strings.ToUpper(mode))
	result.WriteString(ansi.SetBgColor(theme.StatusModeBg))
	result.WriteString(ansi.SetFgColor(theme.StatusFg))
	result.WriteString(ansi.Bold)
	result.WriteString(modeText)
	result.WriteString(ansi.ResetFormat)

	// file segment
	fileText := formatFileSegment(fileName, modified)
	result.WriteString(ansi.SetBgColor(theme.StatusFileBg))
	result.WriteString(ansi.SetFgColor(theme.StatusFg))
	result.WriteString(fileText)
	result.WriteString(ansi.ResetFormat)

	// position segment (rendered at the right end)
	posText := fmt.Sprintf(" %d,%d  %d%% ", row, col, percentage)

	// calculate used visible width so far (strip ANSI codes before measuring)
	usedWidth := visibleLen(modeText) + visibleLen(fileText)

	// language segment if there's space
	lang := syntax.LanguageName(fileName)
	langText := ""
	if lang != "" && usedWidth+len(lang)+3 < width-len(posText)-5 {
		langText = fmt.Sprintf(" %s ", lang)
		result.WriteString(ansi.SetBgColor(theme.StatusLangBg))
		result.WriteString(ansi.SetFgColor(theme.StatusFg))
		result.WriteString(langText)
		result.WriteString(ansi.ResetFormat)
		usedWidth += len(langText)
	}

	// padding between segments and position
	padding := width - usedWidth - len(posText)
	if padding > 0 {
		result.WriteString(ansi.SetBgColor(theme.StatusFileBg))
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(ansi.ResetFormat)
	}

	// position segment
	result.WriteString(ansi.SetBgColor(theme.StatusPosBg))
	result.WriteString(ansi.SetFgColor(theme.StatusFg))
	result.WriteString(posText)
	result.WriteString(ansi.ResetFormat)

	return result.String()
}

// returns an icon for the given mode.
func getModeIcon(mode string) string {
	switch strings.ToLower(mode) {
	case "normal", "nor":
		return "◆"
	case "insert", "ins":
		return "▸"
	case "command", "cmd":
		return ":"
	case "search", "srch":
		return "/"
	case "explore", "expl":
		return "E"
	default:
		return "◆"
	}
}

// formats the file name segment with modified indicator.
func formatFileSegment(fileName string, modified bool) string {
	displayName := fileName
	if fileName == "" {
		displayName = "[No Name]"
	} else {
		displayName = filepath.Base(fileName)
	}

	modifiedStr := ""
	if modified {
		modifiedStr = " [+]"
	}

	return fmt.Sprintf(" %s%s ", displayName, modifiedStr)
}
