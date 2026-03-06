package terminal

import (
	"fmt"
	"io"
	"time"
	"unicode/utf8"
)

// escapeTimeout is the maximum time to wait after ESC to distinguish
// a standalone ESC keypress from the start of an escape sequence.
const escapeTimeout = 50 * time.Millisecond

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

// Key represent a single key event
type Key struct {
	Type KeyType
	Rune rune
	Ctrl bool // whether ctrl was held
	Alt  bool // whether alt was held
}

// inputReader reads bytes from an io.Reader via a background goroutine,
// providing both blocking and timeout-based byte reads.
type inputReader struct {
	ch chan byte
}

func newInputReader(r io.Reader) *inputReader {
	ir := &inputReader{ch: make(chan byte, 256)}
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := r.Read(buf)
			if err != nil || n == 0 {
				close(ir.ch)
				return
			}
			ir.ch <- buf[0]
		}
	}()
	return ir
}

// readByte reads a single byte, blocking until one is available.
func (ir *inputReader) readByte() (byte, error) {
	b, ok := <-ir.ch
	if !ok {
		return 0, io.EOF
	}
	return b, nil
}

// readByteTimeout reads a single byte with a timeout.
// Returns the byte and true if successful, or 0 and false on timeout.
func (ir *inputReader) readByteTimeout(timeout time.Duration) (byte, bool) {
	select {
	case b, ok := <-ir.ch:
		if !ok {
			return 0, false
		}
		return b, true
	case <-time.After(timeout):
		return 0, false
	}
}

// ReadKey reads a single key press from the terminal input.
func (t *Terminal) ReadKey() (*Key, error) {
	b, err := t.input.readByte()
	if err != nil {
		return nil, err
	}

	// handle escape sequences
	if b == 0x1b { // ESC
		return parseEscapeSequence(t.input)
	}

	// handle ASCII control characters (bytes 0x01–0x1f, excluding 0x1b handled above)
	// e.g. 0x0d = Enter, 0x03 = Ctrl+C, 0x12 = Ctrl+R, 0x09 = Tab
	if b < 0x20 {
		return parseControlChar(b)
	}

	// handle backspace (del)
	if b == 0x7f {
		return &Key{Type: KeyBackspace}, nil
	}

	// handle regular characters (can be UTF-8)
	return parseUTF8(t.input, b)
}

func parseEscapeSequence(ir *inputReader) (*Key, error) {
	// Wait briefly for a follow-up byte to distinguish a standalone
	// ESC keypress from the start of a multi-byte escape sequence.
	b1, ok := ir.readByteTimeout(escapeTimeout)
	if !ok {
		// Timeout — user pressed ESC alone
		return &Key{Type: KeyEscape}, nil
	}

	// CSI sequences: ESC [ ...
	if b1 == '[' {
		b2, err := ir.readByte()
		if err != nil {
			return &Key{Type: KeyEscape}, nil
		}

		switch b2 {
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
			if b3, err := ir.readByte(); err == nil && b3 == '~' {
				return &Key{Type: KeyPageUp}, nil
			}
		case '6':
			// Page Down (ESC [ 6 ~)
			if b3, err := ir.readByte(); err == nil && b3 == '~' {
				return &Key{Type: KeyPageDown}, nil
			}
		case '3':
			// Delete (ESC [ 3 ~)
			if b3, err := ir.readByte(); err == nil && b3 == '~' {
				return &Key{Type: KeyDelete}, nil
			}
		}
	}

	// ALT + key sequences (ESC followed by a regular key)
	if b1 >= 'a' && b1 <= 'z' {
		return &Key{
			Type: KeyRune,
			Rune: rune(b1),
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
func parseUTF8(ir *inputReader, firstByte byte) (*Key, error) {
	// single byte ASCII
	if firstByte < 0x80 {
		return &Key{
			Type: KeyRune,
			Rune: rune(firstByte),
		}, nil
	}

	// multi-byte UTF-8 character
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

	// build a correctly-ordered buffer: firstByte at index 0, then the rest
	buf := make([]byte, size)
	buf[0] = firstByte

	for i := 1; i < size; i++ {
		b, err := ir.readByte()
		if err != nil {
			return nil, fmt.Errorf("incomplete UTF-8 character")
		}
		buf[i] = b
	}

	ch, _ := utf8.DecodeRune(buf)
	if ch == utf8.RuneError {
		return nil, fmt.Errorf("invalid UTF-8 sequence")
	}

	return &Key{Type: KeyRune, Rune: ch}, nil
}
