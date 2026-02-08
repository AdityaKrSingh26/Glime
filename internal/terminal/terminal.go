package terminal

import (
	"fmt"
	"io"
	"unicode/utf8"
)

type KeyType int

const (
	KeyRune KeyType = iota
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyBackspace
	KeyDelete
	KeyEnter
	KeyEscape
	KeyTab
	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd
	KeyCtrl // For Ctrl+key combinations
)

// Key reprsent a single key event
type Key struct {
	Type KeyType
	Rune rune
	Ctrl bool // whether ctrl was held
	Alt  bool // whether alt was held
}

// ReadKey reads a single key press from reader
func ReadKey(r io.Reader) (*Key, error) {
	buff := make([]byte, 1)

	// read first byte
	n, err := r.Read(buff)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("no data to read")
	}

	b := buff[0]

	// handle escape sequences
	if b == 0x1b { // ESC
		return parseEscapeSequence(r)
	}

	// handle control characters
	if b == 0x20 {
		return parseControlChar(b)
	}

	// handle backspace (del)
	if b == 0x7f {
		return &Key{Type: KeyBackspace}, nil
	}

	// handle regular characters (can be UTF-8)
	return parseUTF8(r, b)
}

func parseEscapeSequence(r io.Reader) (*Key, error) {
	buff := make([]byte, 2)
	n, err := r.Read(buff)

	// timeout or just ESC
	if err != nil || n == 0 {
		return &Key{Type: KeyEscape}, nil
	}

	// check for arrow keys and other sequencesif
	if buff[0] == '[' {
		if n > 1 {
			switch buff[1] {
			case 'A':
				return &Key{Type: KeyArrowUp}, nil
			case 'B':
				return &Key{Type: KeyArrowDown}, nil
			case 'C':
				return &Key{Type: KeyArrowRight}, nil
			case 'D':
				return &Key{Type: KeyArrowLeft}, nil
			case 'H':
				return &Key{Type: KeyHome}, nil
			case 'F':
				return &Key{Type: KeyEnd}, nil
			case '5':
				// Page Up (ESC [ 5 ~)
				if n, _ := r.Read(buff[:1]); n > 0 && buff[0] == '~' {
					return &Key{Type: KeyPageUp}, nil
				}
			case '6':
				// Page Down (ESC [ 6 ~)
				if n, _ := r.Read(buff[:1]); n > 0 && buff[0] == '~' {
					return &Key{Type: KeyPageDown}, nil
				}
			case '3':
				// Delete (ESC [ 3 ~)
				if n, _ := r.Read(buff[:1]); n > 0 && buff[0] == '~' {
					return &Key{Type: KeyDelete}, nil
				}
			}
		}
	}

	// ALT + key sequences (ESC followed by a regular key)
	if buff[0] >= 'a' && buff[0] <= 'z' {
		return &Key{
			Type: KeyRune,
			Rune: rune(buff[0]),
			Alt:  true,
		}, nil
	}

	// unknown escape sequence, return ESC
	return &Key{Type: KeyEscape}, nil
}

func parseControlChar(b byte) (*Key, error) {
	switch b {
	case 0x0d: // Ctrl+M (Enter)
		return &Key{Type: KeyEnter}, nil
	case 0x08: // Ctrl+H (Backspace)
		return &Key{Type: KeyBackspace}, nil
	case 0x03: // Ctrl+C (Interrupt)
		return &Key{Type: KeyCtrl, Rune: 'c'}, nil
	case 0x1b: // Ctrl+[ (Escape)
		return &Key{Type: KeyCtrl, Rune: '['}, nil
	case 0x09: // Ctrl+I (Tab)
		return &Key{Type: KeyTab}, nil
	default:
		if b >= 0x01 && b <= 0x1a {
			return &Key{
				Type: KeyCtrl,
				Rune: rune('a' + b - 1),
				Ctrl: true,
			}, nil
		}
		return &Key{
			Type: KeyRune,
			Rune: rune(b),
		}, nil
	}
}

// parseUTF8 reads the remaining bytes of a UTF-8 character and returns a Key
func parseUTF8(r io.Reader, firstByte byte) (*Key, error) {

	// single byte ASCII
	if firstByte < 0x80 {
		return &Key{
			Type: KeyRune,
			Rune: rune(firstByte),
		}, nil
	}

	// multi-byte UTF-8 character
	buff := make([]byte, 4)

	// determine how many bytes we need
	var size int
	if firstByte>>5 == 0b110 {
		size = 2
	} else if firstByte>>4 == 0b1110 {
		size = 3
	} else if firstByte>>3 == 0b11110 {
		size = 4
	} else {
		return nil, fmt.Errorf("invalid UTF-8 start byte: %x", firstByte)
	}

	// read remaining bytes
	remaining := make([]byte, size-1)
	n, err := r.Read(remaining)
	if err != nil {
		return nil, err
	}
	if n != size-1 {
		return nil, fmt.Errorf("incomplete UTF-8 character")
	}
	buff = append(buff, remaining...)

	ch, _ := utf8.DecodeRune(buff)
	if ch == utf8.RuneError {
		return nil, fmt.Errorf("invalid UTF-8 sequence")
	}

	return &Key{
		Type: KeyRune,
		Rune: ch,
	}, nil
}
