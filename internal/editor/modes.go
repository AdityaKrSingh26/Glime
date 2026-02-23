package editor

type Mode int

const (
	ModeNormal  Mode = iota // default mode for navigation and commands
	ModeInsert              // mode for inserting text
	ModeCommand             // mode for entering commands (eg :w , :q)
	ModeSearch              // mode for incremental search (ed /, ?)
)

// returns the string representation of modes
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeCommand:
		return "COMMAND"
	case ModeSearch:
		return "SEARCH"
	default:
		return "UNKNOWN"
	}
}

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
	default:
		return "???"
	}
}
