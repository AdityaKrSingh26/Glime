package editor

type Mode int

const (
	ModeNormal  Mode = iota // default mode for navigation and commands
	ModeInsert              // mode for inserting text
	ModeCommand             // mode for entering commands (eg :w , :q)
	ModeSearch              // mode for incremental search (ed /, ?)
	ModeExplore             // mode for file explorer (netrw-style)
)

// represent a short string representation for the status bar
func (m Mode) ShortString() string {
	switch m {
	case ModeNormal:
		return "NOR"
	case ModeInsert:
		return "INS"
	case ModeCommand:
		return "CMD"
	case ModeSearch:
		return "SRCH"
	case ModeExplore:
		return "EXPL"
	default:
		return "???"
	}
}
