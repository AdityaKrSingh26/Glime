package editor

import "unicode"

// finds the position after the next word boundary
// return the (row, col) of the start of the next word (rune-indexed)
func (e *Editor) findWordEnd(row, col int) (int, int) {
	lines := e.buffer.GetLines()
	if row >= len(lines) {
		return row, col
	}

	runes := []rune(lines[row])

	// skip over empty lines at current position
	for col >= len(runes) {
		if row < len(lines)-1 {
			row++
			col = 0
			runes = []rune(lines[row])
		} else {
			return row, col
		}
	}

	// determine current character class
	ch := runes[col]
	isWord := isWordChar(ch)
	isPunct := !isWord && !unicode.IsSpace(ch)

	// skip current word/punctuation
	for col < len(runes) {
		c := runes[col]
		if isWord && !isWordChar(c) {
			break
		}
		if isPunct && (isWordChar(c) || unicode.IsSpace(c)) {
			break
		}
		if unicode.IsSpace(ch) && !unicode.IsSpace(c) {
			break
		}
		col++
	}

	for col < len(runes) && unicode.IsSpace(runes[col]) {
		col++
	}

	// end of line, go to start of next line
	if col >= len(runes) && row < len(lines)-1 {
		return row + 1, 0
	}

	return row, col
}

// finds the position of the previous word start (rune-indexed)
func (e *Editor) findWordStart(row, col int) (int, int) {
	lines := e.buffer.GetLines()
	if row >= len(lines) {
		return row, col
	}

	// if at start of line, move to end of previous line
	if col == 0 {
		if row > 0 {
			prevRunes := []rune(lines[row-1])
			if len(prevRunes) == 0 {
				return row - 1, 0
			}
			return row - 1, len(prevRunes) - 1
		}
		return 0, 0
	}

	runes := []rune(lines[row])
	col-- // move back one

	// skip trailing whitespace to get to the previous word
	for col > 0 && unicode.IsSpace(runes[col]) {
		col--
	}

	if col == 0 {
		return row, 0
	}

	// determine current character class
	ch := runes[col]
	isWord := isWordChar(ch)

	// skip same class backward
	for col > 0 {
		prev := runes[col-1]
		if isWord && !isWordChar(prev) {
			break
		}
		if !isWord && (isWordChar(prev) || unicode.IsSpace(prev)) {
			break
		}
		col--
	}

	return row, col
}

// deletes count lines starting from the current cursor row.
func (e *Editor) deleteLines(count int) error {
	row := e.cursor.Row()
	lines := e.buffer.GetLines()

	if count > len(lines)-row {
		count = len(lines) - row
	}

	// collect lines for yank register
	var yanked string
	for i := 0; i < count; i++ {
		if i > 0 {
			yanked += "\n"
		}
		yanked += lines[row+i]
	}
	e.register = Register{
		Content: yanked,
		Type:    RegisterLine,
	}

	// record undo actions (reverse order so undo re-inserts bottom-up)
	e.undoMgr.BeginGroup()
	for i := count - 1; i >= 0; i-- {
		lineText := lines[row+i]
		e.undoMgr.Record(Action{
			Type:      ActionDeleteLine,
			Row:       row + i,
			Text:      lineText,
			CursorRow: e.cursor.Row(),
			CursorCol: e.cursor.Col(),
		})
	}

	// perform deletions
	for i := 0; i < count; i++ {
		e.buffer.DeleteLine(row)
	}
	e.undoMgr.EndGroup()

	// adjust cursor position
	if row >= e.buffer.NumLines() {
		row = e.buffer.NumLines() - 1
	}
	if row < 0 {
		row = 0
	}
	e.cursor.MoveTo(row, 0, e.buffer)

	return nil
}

// deleteWord deletes from cursor to next word boundary (count times). Rune-indexed.
func (e *Editor) deleteWord(count int) error {
	row := e.cursor.Row()
	col := e.cursor.Col()

	endRow, endCol := row, col
	for i := 0; i < count; i++ {
		endRow, endCol = e.findWordEnd(endRow, endCol)
	}

	if endRow == row {
		// same line deletion
		line, _ := e.buffer.GetLine(row)
		runes := []rune(line)
		if endCol > len(runes) {
			endCol = len(runes)
		}
		deleted := string(runes[col:endCol])
		e.register = Register{
			Content: deleted,
			Type:    RegisterChar,
		}

		newLine := string(runes[:col]) + string(runes[endCol:])
		e.undoMgr.Record(Action{
			Type:      ActionSetLine,
			Row:       row,
			Text:      newLine,
			PrevText:  line,
			CursorRow: e.cursor.Row(),
			CursorCol: e.cursor.Col(),
		})
		e.buffer.SetLine(row, newLine)
	} else {
		// multi-line deletion
		e.undoMgr.BeginGroup()

		lines := e.buffer.GetLines()
		var deleted string

		// capture text being deleted (rune-safe)
		firstLine := lines[row]
		firstRunes := []rune(firstLine)
		deleted = string(firstRunes[col:])
		for r := row + 1; r < endRow; r++ {
			deleted += "\n" + lines[r]
		}
		if endRow < len(lines) {
			endRunes := []rune(lines[endRow])
			deleted += "\n" + string(endRunes[:endCol])
		}
		e.register = Register{Content: deleted, Type: RegisterChar}

		// Merge: keep beginning of first line + end of last line
		newLine := string(firstRunes[:col])
		if endRow < len(lines) {
			endRunes := []rune(lines[endRow])
			newLine += string(endRunes[endCol:])
		}

		// Record SetLine FIRST, then DeleteLine entries (reverse order)
		// On undo (applied in reverse): lines get re-inserted first, then first line restored
		e.undoMgr.Record(Action{
			Type:      ActionSetLine,
			Row:       row,
			Text:      newLine,
			PrevText:  firstLine,
			CursorRow: e.cursor.Row(),
			CursorCol: e.cursor.Col(),
		})
		for r := endRow; r > row; r-- {
			e.undoMgr.Record(Action{
				Type:      ActionDeleteLine,
				Row:       r,
				Text:      lines[r],
				CursorRow: e.cursor.Row(),
				CursorCol: e.cursor.Col(),
			})
		}
		e.undoMgr.EndGroup()

		// perform deletions
		e.buffer.SetLine(row, newLine)
		for r := endRow; r > row; r-- {
			e.buffer.DeleteLine(r)
		}
	}

	// Clamp cursor
	e.cursor.MoveTo(e.cursor.Row(), col, e.buffer)
	return nil
}

// deletes from cursor to end of line (rune-indexed).
func (e *Editor) deleteToLineEnd() error {
	row := e.cursor.Row()
	col := e.cursor.Col()
	line, _ := e.buffer.GetLine(row)
	runes := []rune(line)

	if col >= len(runes) {
		return nil
	}

	deleted := string(runes[col:])
	e.register = Register{Content: deleted, Type: RegisterChar}

	newLine := string(runes[:col])
	e.undoMgr.Record(Action{
		Type:      ActionSetLine,
		Row:       row,
		Text:      newLine,
		PrevText:  line,
		CursorRow: row,
		CursorCol: col,
	})
	e.buffer.SetLine(row, newLine)

	// Clamp cursor
	e.cursor.MoveTo(row, col, e.buffer)
	return nil
}

// returns true if the rune is a word character (alphanumeric or _).
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
