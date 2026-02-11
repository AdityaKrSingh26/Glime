package terminal

// raw mode -> terminal suspends all its special behaviors (eg: CTRL+C-> cancel)

import (
	"fmt"
	"golang.org/x/term"
)

// In raw mode:
//   - Input is available immediately (no buffering)
//   - No echo of typed characters
//   - Special keys (Ctrl+C, Ctrl+Z) are not processed by the terminal
//   - We receive raw bytes for all input
// The original terminal state is saved so it can be restored later.
// Always call DisableRawMode() before the program exits.
func (t *Terminal) EnableRawMode() error {
	if t.originalState != nil {
		return fmt.Errorf("raw mode already enabled")
	}

	// save orignal terminal state
	state, err := term.MakeRaw(t.fd)
	if err!=nil {
		return fmt.Errorf("failed to enter raw mode: %w", err)
	}

	t.originalState = state
	return nil
}

// DisableRawMode restores the terminal to its original state.
// This should always be called before program exit, typically in a defer.
func (t *Terminal) DisableRawMode() error {
	if t.originalState != nil {
		return nil // already in raw mode
	}

	if err := term.Restore(t.fd, t.originalState); err != nil {
		return fmt.Errorf("failed to disable raw mode: %w", err)
	}

	t.originalState = nil
	return nil
}

// check if in raw mode
func (t *terminal) IsRawMode() bool {
	return t.originalState != nil
}