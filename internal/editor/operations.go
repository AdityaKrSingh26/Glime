package editor

import "unicode"

// finds the position after the next word boundary
// return the (row, col) of the start of the next word
func (e *Editor) findWordEnd(row, col int) (int, int) {
	lines := e.buffer.GetLines()
	if row >= len(lines) {
		return row, col
	}

	line := lines[row]

	// if at end of line, move to next line
	if col >= len(line) {
		if row < len(lines)-1 {
			return row + 1, 0
		}
		return row, col
	}

	// determine current character class
	ch := rune(line[col])
	isWord := isWordChar(ch)
	isPunct := !isWord && !unicode.IsSpace(ch)

	// skip current word/punctuation
	for col < len(line) {
		c := rune(line[col])
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

	for col < len(line) && unicode.IsSpace(rune(line[col])) {
		col++
	}

	// end of line, go to start of next line
	if col >= len(line) && row < len(lines)-1 {
		return row + 1, 0
	}

	return row, col
}

// finds the position of the previous word start
func (e *Editor) findWordStart(row, col int) (int, int) {
	lines := e.buffer.GetLines()
	if row >= len(lines) {
		return row, col
	}

	// if at start of line, move to end of previous line
	if col == 0 {
		if row > 0 {
			prevLen := len(lines[row-1])
			if prevLen == 0 {
				return row - 1, 0
			}
			return row - 1, prevLen - 1
		}
		return 0, 0
	}

	line := lines[row]
	col-- // move back one

	// skip trailing whitespace to get to the previous word
	for col > 0 && unicode.IsSpace(rune(line[col])) {
		col--
	}

	if col == 0 {
		// col 0 may itself be the start of a word — let the caller handle it
		return row, 0
	}

	// determine current character class
	ch := rune(line[col])
	isWord := isWordChar(ch)

	// skip same class backward
	for col > 0 {
		prev := rune(line[col-1])
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

// deleteWord deletes from cursor to next word boundary (count times).
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
		if endCol > len(line) {
			endCol = len(line)
		}
		deleted := line[col:endCol]
		e.register = Register{
			Content: deleted,
			Type:    RegisterChar,
		}

		e.undoMgr.Record(Action{
			Type:      ActionSetLine,
			Row:       row,
			Text:      line[:col] + line[endCol:],
			PrevText:  line,
			CursorRow: e.cursor.Row(),
			CursorCol: e.cursor.Col(),
		})
		e.buffer.SetLine(row, line[:col]+line[endCol:])
	} else {
		// multi-line deletion: delete from col to end of current line,
		// then delete intermediate lines, then delete start of target line
		e.undoMgr.BeginGroup()

		lines := e.buffer.GetLines()
		var deleted string

		// capture text being deleted
		firstLine := lines[row]
		deleted = firstLine[col:]
		for r := row + 1; r < endRow; r++ {
			deleted += "\n" + lines[r]
		}
		if endRow < len(lines) {
			deleted += "\n" + lines[endRow][:endCol]
		}
		e.register = Register{Content: deleted, Type: RegisterChar}

		// Merge: keep beginning of first line + end of last line
		newLine := firstLine[:col]
		if endRow < len(lines) {
			newLine += lines[endRow][endCol:]
		}

		// record undo for each line being removed (reverse order)
		for r := endRow; r > row; r-- {
			e.undoMgr.Record(Action{
				Type:      ActionDeleteLine,
				Row:       r,
				Text:      lines[r],
				CursorRow: e.cursor.Row(),
				CursorCol: e.cursor.Col(),
			})
		}
		e.undoMgr.Record(Action{
			Type:      ActionSetLine,
			Row:       row,
			Text:      newLine,
			PrevText:  firstLine,
			CursorRow: e.cursor.Row(),
			CursorCol: e.cursor.Col(),
		})
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

// deletes from cursor to end of line.
func (e *Editor) deleteToLineEnd() error {
	row := e.cursor.Row()
	col := e.cursor.Col()
	line, _ := e.buffer.GetLine(row)

	if col >= len(line) {
		return nil
	}

	deleted := line[col:]
	e.register = Register{Content: deleted, Type: RegisterChar}

	e.undoMgr.Record(Action{
		Type:      ActionSetLine,
		Row:       row,
		Text:      line[:col],
		PrevText:  line,
		CursorRow: row,
		CursorCol: col,
	})
	e.buffer.SetLine(row, line[:col])

	// Clamp cursor
	e.cursor.MoveTo(row, col, e.buffer)
	return nil
}

// returns true if the rune is a word character (alphanumeric or _).
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
