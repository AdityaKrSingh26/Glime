package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AdityaKrSingh26/Glime/internal/file"
)

// execute command in command mode, command start with ":" and executed when enter is pressed
func (e *Editor) executeCommand(cmd string) error {
	// remove leading ":"
	cmd = strings.TrimPrefix(cmd, ":")
	cmd = strings.TrimSpace(cmd)

	if cmd == "" {
		return nil
	}

	parts := strings.Fields(cmd)
	command := parts[0]

	switch command {
	case "q":
		return e.commandQuit(false)
	case "q!":
		return e.commandQuit(true)
	case "w":
		return e.commandWrite()
	case "wq":
		return e.commandWriteQuit()
	case "x":
		return e.commandWriteQuit() // Same as :wq
	default:
		// Check if it's a line number (e.g., :42)
		if lineNum, err := strconv.Atoi(command); err == nil {
			return e.commandGotoLine(lineNum)
		}
		e.setMessage(fmt.Sprintf("Unknown command: %s", command))
	}

	return nil
}

// quits the editor, if force is false, it checks for unsaved changes.
func (e *Editor) commandQuit(force bool) error {
	if !force && e.buffer.IsModified() {
		e.setMessage("No write since last change (use :q! to override)")
		return nil
	}

	e.shouldQuit = true
	return nil
}

// saves the buffer to disk.
func (e *Editor) commandWrite() error {
	filePath := e.buffer.FilePath()

	if filePath == "" {
		e.setMessage("No file name")
		return nil
	}

	// Back up existing file before overwriting
	if err := file.Backup(filePath); err != nil {
		e.setMessage(fmt.Sprintf("Backup failed: %v", err))
		return err
	}

	// Save the file
	if err := file.Save(filePath, e.buffer.GetLines()); err != nil {
		e.setMessage(fmt.Sprintf("Error writing file: %v", err))
		return err
	}

	e.buffer.SetModified(false)
	e.setMessage(fmt.Sprintf("\"%s\" %dL written", e.buffer.FileName(), e.buffer.NumLines()))
	return nil
}

func (e *Editor) commandWriteQuit() error {
	if err := e.commandWrite(); err != nil {
		return err
	}

	// quit only if save was successful
	if !e.buffer.IsModified() {
		e.shouldQuit = true
	}

	return nil
}

// moves the cursor to the specified line number (1-indexed).
func (e *Editor) commandGotoLine(lineNum int) error {
	// Convert to 0-indexed
	lineNum--

	if lineNum < 0 {
		lineNum = 0
	}
	if lineNum >= e.buffer.NumLines() {
		lineNum = e.buffer.NumLines() - 1
	}

	e.cursor.MoveTo(lineNum, 0, e.buffer)
	e.setMessage(fmt.Sprintf("Line %d", lineNum+1))
	return nil
}
