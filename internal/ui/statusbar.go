package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AdityaKrSingh26/Glime/internal/syntax"
	"github.com/AdityaKrSingh26/Glime/pkg/ansi"
)

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

	// Mode segment
	icon := getModeIcon(mode)
	modeText := fmt.Sprintf(" %s %s ", icon, strings.ToUpper(mode))
	result.WriteString(ansi.SetBgColor(theme.StatusModeBg))
	result.WriteString(ansi.SetFgColor(theme.StatusFg))
	result.WriteString(ansi.Bold)
	result.WriteString(modeText)
	result.WriteString(ansi.ResetFormat)

	// File segment
	fileText := formatFileSegment(fileName, modified)
	result.WriteString(ansi.SetBgColor(theme.StatusFileBg))
	result.WriteString(ansi.SetFgColor(theme.StatusFg))
	result.WriteString(fileText)
	result.WriteString(ansi.ResetFormat)

	// Position segment (rendered at the right end)
	posText := fmt.Sprintf(" %d,%d  %d%% ", row, col, percentage)

	// Calculate used width so far
	usedWidth := len(modeText) + len(fileText)

	// Language segment if there's space
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

	// Padding between segments and position
	padding := width - usedWidth - len(posText)
	if padding > 0 {
		result.WriteString(ansi.SetBgColor(theme.StatusFileBg))
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(ansi.ResetFormat)
	}

	// Position segment
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
