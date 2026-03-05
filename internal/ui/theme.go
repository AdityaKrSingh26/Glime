package ui

// defines the color scheme for the editor using 256-color palette.
type Theme struct {
	// Syntax highlighting colors (256-color indices)
	Keyword  int
	String   int
	Comment  int
	Number   int
	Function int
	Type     int
	Operator int // Operators like +, -, *, /
	Builtin  int // Built-in functions/types

	// UI element colors
	LineNumber        int // Line nos in gutter
	CurrentLineNumber int // Current line no highlight
	Gutter            int // Gutter background
	EmptyLine         int // Tilde for empty lines
	Border            int // Border elements

	// Status bar colors
	StatusModeBg int // Mode segment background
	StatusFileBg int // File info segment background
	StatusLangBg int // Language segment background
	StatusPosBg  int // Position segment background
	StatusFg     int // Status bar foreground text

	// Search and bracket matching
	SearchHighlight int // Background color for search matches
	BracketMatch    int // Background color for matching brackets
}

// returns the "Glime Modern Dark" theme with vibrant colors.
func DefaultTheme() Theme {
	return Theme{
		// Syntax colors
		Keyword:  204, // Bright pink/magenta
		String:   114, // Light green
		Comment:  243, // Dim gray
		Number:   215, // Peach/orange
		Function: 117, // Sky blue
		Type:     141, // Purple
		Operator: 208, // Orange
		Builtin:  216, // Light peach

		// UI colors
		LineNumber:        240, // Dark gray
		CurrentLineNumber: 220, // Golden yellow
		Gutter:            235, // Very dark gray background
		EmptyLine:         239, // Medium dark gray
		Border:            238, // Border gray

		// Status bar colors
		StatusModeBg: 24,  // Deep blue
		StatusFileBg: 236, // Very dark gray
		StatusLangBg: 238, // Slightly lighter gray
		StatusPosBg:  240, // Medium gray
		StatusFg:     255, // Bright white

		// Search and bracket matching
		SearchHighlight: 226, // Yellow background
		BracketMatch:    240, // Medium gray background
	}
}
