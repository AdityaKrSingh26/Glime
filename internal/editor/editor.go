package editor

// this package provide the main editor state machine and event loop

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AdityaKrSingh26/Glime/internal/buffer"
	"github.com/AdityaKrSingh26/Glime/internal/cursor"
	"github.com/AdityaKrSingh26/Glime/internal/file"
	"github.com/AdityaKrSingh26/Glime/internal/input"
	"github.com/AdityaKrSingh26/Glime/internal/terminal"
	"github.com/AdityaKrSingh26/Glime/internal/ui"
)

// represents the main editor state
type Editor struct {
	terminal *terminal.Terminal
	buffer   *buffer.Buffer
	cursor   *cursor.Cursor
	renderer *ui.Renderer

	mode        Mode
	message     string
	messageTime time.Time
	commandBuf  string // Buffer for command mode input
	shouldQuit  bool

	undoMgr   *UndoManager   // Undo/Redo
	pending   PendingCommand // Multi-key commands
	register  Register       // Copy/Paste
	search    SearchState    // Search
	searchBuf string         // Input buffer for search mode
}

func New() (*Editor, error) {
	term, err := terminal.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create new terminal: %w", err)
	}

	buf := buffer.New()
	cur := cursor.New()
	renderer := ui.NewRenderer(term)

	return &Editor{
		terminal:   term,
		buffer:     buf,
		cursor:     cur,
		renderer:   renderer,
		mode:       ModeNormal,
		message:    "Glime editor - Type :q to quit",
		commandBuf: "",
		shouldQuit: false,
		undoMgr:    NewUndoManager(1000),
	}, nil
}

// load file into editor
func (e *Editor) LoadFile(filePath string) error {
	if !file.Exists(filePath) {
		// file does not exist create a buffer with this path
		e.buffer.SetFilePath(filePath)
		e.renderer.SetLanguage(filePath)
		e.setMessage(fmt.Sprintf("\"%s\" [New File]", filePath))
		return nil
	}

	lines, err := file.Load(filePath)
	if err != nil {
		return fmt.Errorf("Error loading file into editor: %w", err)
	}

	e.buffer = buffer.NewFromLines(lines, filePath)
	e.renderer.SetLanguage(filePath)
	e.setMessage(fmt.Sprintf("\"%s\" %dL", filePath, e.buffer.NumLines()))
	return nil
}

// start the editor event loop
func (e *Editor) Run() error {
	// enable raw mode
	if err := e.terminal.EnableRawMode(); err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	defer e.terminal.DisableRawMode()

	// enable alternate buffer
	if err := e.terminal.EnableAlternateBuffer(); err != nil {
		return fmt.Errorf("failed to enable alternate buffer: %w", err)
	}
	defer e.terminal.DisableAlternateBuffer()

	// watch for terminal resize (SIGWINCH)
	e.terminal.WatchResize()

	// hide cursor during setup
	if err := e.terminal.HideCursor(); err != nil {
		return err
	}

	// clear screen
	if err := e.terminal.Clear(); err != nil {
		return err
	}

	// main event loop
	for !e.shouldQuit {
		// update scroll to make sure cursor is visible
		e.updateScroll()

		// render a new screen
		view := e.buildView()
		if err := e.renderer.Render(view); err != nil {
			return fmt.Errorf("render error: %w", err)
		}

		// read key input
		key, err := input.ReadKey(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read key: %w", err)
		}

		// process the key
		if err := e.processKey(key); err != nil {
			return fmt.Errorf("key processing error: %w", err)
		}
	}

	return nil
}

// to handle a key press based on the current mode.
func (e *Editor) processKey(key *input.Key) error {
	switch e.mode {
	case ModeNormal:
		return e.processNormalMode(key)
	case ModeInsert:
		return e.processInsertMode(key)
	case ModeCommand:
		return e.processCommandMode(key)
	case ModeSearch:
		return e.processSearchMode(key)
	}
	return nil
}

// handles keys in Normal mode with multi-key support.
func (e *Editor) processNormalMode(key *input.Key) error {
	// Step 1: Handle non-rune keys first (arrows, page, ctrl, escape)
	switch key.Type {
	case input.KeyArrowLeft:
		e.pending.Reset()
		e.cursor.MoveLeft(e.buffer)
		return nil
	case input.KeyArrowRight:
		e.pending.Reset()
		e.cursor.MoveRight(e.buffer)
		return nil
	case input.KeyArrowUp:
		e.pending.Reset()
		e.cursor.MoveUp(e.buffer)
		return nil
	case input.KeyArrowDown:
		e.pending.Reset()
		e.cursor.MoveDown(e.buffer)
		return nil
	case input.KeyPageUp:
		e.pending.Reset()
		e.cursor.PageUp(e.buffer, e.terminal.Height()-2)
		return nil
	case input.KeyPageDown:
		e.pending.Reset()
		e.cursor.PageDown(e.buffer, e.terminal.Height()-2)
		return nil
	case input.KeyHome:
		e.pending.Reset()
		e.cursor.MoveToLineStart()
		return nil
	case input.KeyEnd:
		e.pending.Reset()
		e.cursor.MoveToLineEndNormal(e.buffer)
		return nil
	case input.KeyEscape:
		e.pending.Reset()
		return nil
	case input.KeyCtrl:
		e.pending.Reset()
		switch key.Rune {
		case 'c':
			e.shouldQuit = true
		case 'r':
			e.redo()
		}
		return nil
	}

	if key.Type != input.KeyRune {
		return nil
	}

	ch := key.Rune

	// Step 2: Accumulate count prefix
	if ch >= '1' && ch <= '9' {
		e.pending.AccumulateDigit(int(ch - '0'))
		return nil
	}
	if ch == '0' && e.pending.HasCount {
		e.pending.AccumulateDigit(0)
		return nil
	}

	// Step 3: Operator-pending mode (d, y, g wait for second key)
	if e.pending.Operator == 0 {
		switch ch {
		case 'd', 'y', 'g':
			e.pending.Operator = ch
			return nil
		}
	}

	// Step 4: Execute command
	return e.executeNormalCommand(ch)
}

// dispatches a normal mode command (possibly with pending operator).
func (e *Editor) executeNormalCommand(ch rune) error {
	count := e.pending.EffectiveCount()
	op := e.pending.Operator
	defer e.pending.Reset()

	// Operator + motion combinations
	if op != 0 {
		switch op {
		case 'd':
			switch ch {
			case 'd':
				return e.deleteLines(count)
			case 'w':
				return e.deleteWord(count)
			case '$':
				return e.deleteToLineEnd()
			default:
				return nil // unknown motion, ignore
			}
		case 'y':
			switch ch {
			case 'y':
				return e.yankLines(count)
			case 'w':
				return e.yankWord(count)
			case '$':
				return e.yankToLineEnd()
			default:
				return nil
			}
		case 'g':
			switch ch {
			case 'g':
				e.cursor.MoveToFirstLine()
				return nil
			default:
				return nil
			}
		}
		return nil
	}

	// Simple commands (no operator pending)
	switch ch {
	case 'i':
		e.setMode(ModeInsert)
		e.undoMgr.BeginGroup()
	case ':':
		e.setMode(ModeCommand)
		e.commandBuf = ":"
	case '/':
		e.enterSearchMode(SearchForward)
	case '?':
		e.enterSearchMode(SearchBackward)
	case 'n':
		e.searchNext()
	case 'N':
		e.searchPrev()
	case 'h':
		for i := 0; i < count; i++ {
			e.cursor.MoveLeft(e.buffer)
		}
	case 'j':
		for i := 0; i < count; i++ {
			e.cursor.MoveDown(e.buffer)
		}
	case 'k':
		for i := 0; i < count; i++ {
			e.cursor.MoveUp(e.buffer)
		}
	case 'l':
		for i := 0; i < count; i++ {
			e.cursor.MoveRight(e.buffer)
		}
	case '0':
		e.cursor.MoveToLineStart()
	case '$':
		e.cursor.MoveToLineEndNormal(e.buffer)
	case 'G':
		e.cursor.MoveToLastLine(e.buffer)
	case 'w':
		for i := 0; i < count; i++ {
			row, col := e.findWordEnd(e.cursor.Row(), e.cursor.Col())
			e.cursor.MoveTo(row, col, e.buffer)
		}
	case 'b':
		for i := 0; i < count; i++ {
			row, col := e.findWordStart(e.cursor.Row(), e.cursor.Col())
			e.cursor.MoveTo(row, col, e.buffer)
		}
	case 'x':
		e.deleteCharUnderCursor()
	case 'u':
		e.undo()
	case 'p':
		e.pasteAfter()
	case 'P':
		e.pasteBefore()
	}

	return nil
}

// handles keys in Insert mode.
func (e *Editor) processInsertMode(key *input.Key) error {
	switch key.Type {
	case input.KeyEscape:
		e.undoMgr.EndGroup()
		e.setMode(ModeNormal)

	case input.KeyRune:
		e.insertChar(e.cursor.Row(), e.cursor.Col(), key.Rune)
		e.cursor.MoveRight(e.buffer)

	case input.KeyEnter:
		e.splitLine(e.cursor.Row(), e.cursor.Col())
		e.cursor.MoveDown(e.buffer)
		e.cursor.MoveToLineStart()

	case input.KeyBackspace:
		e.backspace(e.cursor.Row(), e.cursor.Col())

	case input.KeyDelete:
		e.deleteCharAt(e.cursor.Row(), e.cursor.Col())

	case input.KeyArrowLeft:
		e.cursor.MoveLeft(e.buffer)
	case input.KeyArrowRight:
		e.cursor.MoveRight(e.buffer)
	case input.KeyArrowUp:
		e.cursor.MoveUp(e.buffer)
	case input.KeyArrowDown:
		e.cursor.MoveDown(e.buffer)
	}

	return nil
}

func (e *Editor) processCommandMode(key *input.Key) error {
	switch key.Type {
	case input.KeyEscape:
		e.setMode(ModeNormal)
		e.commandBuf = ""

	case input.KeyEnter:
		// Execute the command
		if err := e.executeCommand(e.commandBuf); err != nil {
			e.setMessage(fmt.Sprintf("Error: %v", err))
		}

		// Don't clear message if we're quitting
		if !e.shouldQuit {
			e.setMode(ModeNormal)
		}
		e.commandBuf = ""

	case input.KeyBackspace:
		if len(e.commandBuf) > 1 { // Keep the ':'
			e.commandBuf = e.commandBuf[:len(e.commandBuf)-1]
		}

	case input.KeyRune:
		e.commandBuf += string(key.Rune)
	}

	return nil
}

func (e *Editor) processSearchMode(key *input.Key) error {
	switch key.Type {
	case input.KeyEscape:
		// Cancel search
		e.search.Active = false
		e.search.Pattern = ""
		e.search.Matches = nil
		e.searchBuf = ""
		e.setMode(ModeNormal)

	case input.KeyEnter:
		// Finalize search and jump to nearest match
		e.search.Pattern = e.searchBuf
		e.search.FindAll(e.buffer.GetLines())
		if len(e.search.Matches) > 0 {
			idx := e.search.NextMatch(e.cursor.Row(), e.cursor.Col())
			if idx >= 0 {
				m := e.search.Matches[idx]
				e.search.CurrentIndex = idx
				e.cursor.MoveTo(m.Row, m.ColStart, e.buffer)
			}
			e.setMessage(fmt.Sprintf("/%s [%d matches]", e.search.Pattern, len(e.search.Matches)))
		} else {
			e.setMessage(fmt.Sprintf("Pattern not found: %s", e.searchBuf))
			e.search.Active = false
		}
		e.searchBuf = ""
		e.setMode(ModeNormal)

	case input.KeyBackspace:
		if len(e.searchBuf) > 0 {
			e.searchBuf = e.searchBuf[:len(e.searchBuf)-1]
			// Re-run incremental search
			e.search.Pattern = e.searchBuf
			e.search.FindAll(e.buffer.GetLines())
		}

	case input.KeyRune:
		e.searchBuf += string(key.Rune)
		// Incremental search
		e.search.Pattern = e.searchBuf
		e.search.FindAll(e.buffer.GetLines())
	}

	return nil
}

func (e *Editor) enterSearchMode(dir SearchDirection) {
	e.search.Direction = dir
	e.search.Active = true
	e.search.Matches = nil
	e.searchBuf = ""
	e.setMode(ModeSearch)
}

// searchNext jumps in the original search direction (n key).
func (e *Editor) searchNext() {
	if !e.search.Active || len(e.search.Matches) == 0 {
		e.setMessage("No search pattern")
		return
	}

	var idx int
	if e.search.Direction == SearchForward {
		idx = e.search.NextMatch(e.cursor.Row(), e.cursor.Col())
	} else {
		idx = e.search.PrevMatch(e.cursor.Row(), e.cursor.Col())
	}
	if idx >= 0 {
		m := e.search.Matches[idx]
		e.search.CurrentIndex = idx
		e.cursor.MoveTo(m.Row, m.ColStart, e.buffer)
		dir := "/"
		if e.search.Direction == SearchBackward {
			dir = "?"
		}
		e.setMessage(fmt.Sprintf("%s%s [%d/%d]", dir, e.search.Pattern, idx+1, len(e.search.Matches)))
	}
}

// searchPrev jumps opposite to the original search direction (N key).
func (e *Editor) searchPrev() {
	if !e.search.Active || len(e.search.Matches) == 0 {
		e.setMessage("No search pattern")
		return
	}

	var idx int
	if e.search.Direction == SearchForward {
		idx = e.search.PrevMatch(e.cursor.Row(), e.cursor.Col())
	} else {
		idx = e.search.NextMatch(e.cursor.Row(), e.cursor.Col())
	}
	if idx >= 0 {
		m := e.search.Matches[idx]
		e.search.CurrentIndex = idx
		e.cursor.MoveTo(m.Row, m.ColStart, e.buffer)
		dir := "/"
		if e.search.Direction == SearchBackward {
			dir = "?"
		}
		e.setMessage(fmt.Sprintf("%s%s [%d/%d]", dir, e.search.Pattern, idx+1, len(e.search.Matches)))
	}
}

// --- Undo-aware buffer wrapper methods ---

// records and executes a character insertion.
func (e *Editor) insertChar(row, col int, ch rune) {
	e.undoMgr.Record(Action{
		Type:      ActionInsertChar,
		Row:       row,
		Col:       col,
		Text:      string(ch),
		CursorRow: row,
		CursorCol: col,
	})
	e.buffer.InsertChar(row, col, ch)
}

// records and executes a character deletion.
func (e *Editor) deleteCharAt(row, col int) {
	line, err := e.buffer.GetLine(row)
	if err != nil || col >= len(line) {
		return
	}
	ch := string(line[col])

	e.undoMgr.Record(Action{
		Type:      ActionDeleteChar,
		Row:       row,
		Col:       col,
		Text:      ch,
		CursorRow: row,
		CursorCol: col,
	})
	e.buffer.DeleteChar(row, col)
}

// deletes the char under cursor and yanks it.
func (e *Editor) deleteCharUnderCursor() {
	line, err := e.buffer.GetLine(e.cursor.Row())
	if err != nil || e.cursor.Col() >= len(line) {
		return
	}
	ch := string(line[e.cursor.Col()])
	e.register = Register{Content: ch, Type: RegisterChar}
	e.deleteCharAt(e.cursor.Row(), e.cursor.Col())
}

func (e *Editor) splitLine(row, col int) {
	line, _ := e.buffer.GetLine(row)
	e.undoMgr.Record(Action{
		Type:      ActionSplitLine,
		Row:       row,
		Col:       col,
		PrevText:  line,
		CursorRow: row,
		CursorCol: col,
	})
	e.buffer.SplitLine(row, col)
}

func (e *Editor) joinLines(row int) {
	lineA, _ := e.buffer.GetLine(row)
	lineB, _ := e.buffer.GetLine(row + 1)
	e.undoMgr.Record(Action{
		Type:      ActionJoinLines,
		Row:       row,
		Text:      lineA,
		PrevText:  lineB,
		CursorRow: e.cursor.Row(),
		CursorCol: e.cursor.Col(),
	})
	e.buffer.JoinLines(row)
}

func (e *Editor) backspace(row, col int) {
	if row == 0 && col == 0 {
		return
	}

	if col == 0 {
		// Join with previous line
		prevLen, _ := e.buffer.LineLength(row - 1)
		e.joinLines(row - 1)
		e.cursor.SetPosition(row-1, prevLen)
	} else {
		line, _ := e.buffer.GetLine(row)
		if col > len(line) {
			col = len(line)
		}
		ch := string(line[col-1])
		e.undoMgr.Record(Action{
			Type:      ActionDeleteChar,
			Row:       row,
			Col:       col - 1,
			Text:      ch,
			CursorRow: row,
			CursorCol: col,
		})
		e.buffer.DeleteChar(row, col-1)
		e.cursor.SetPosition(row, col-1)
	}
}

// --- Yank operations ---
// When you "yank" text, it gets copied into the register,
// so you can paste it somewhere else later.

// copies count lines to the register.
func (e *Editor) yankLines(count int) error {
	row := e.cursor.Row()
	lines := e.buffer.GetLines()

	if count > len(lines)-row {
		count = len(lines) - row
	}

	var yanked string
	for i := 0; i < count; i++ {
		if i > 0 {
			yanked += "\n"
		}
		yanked += lines[row+i]
	}
	e.register = Register{Content: yanked, Type: RegisterLine}

	if count == 1 {
		e.setMessage("1 line yanked")
	} else {
		e.setMessage(fmt.Sprintf("%d lines yanked", count))
	}
	return nil
}

// copies from cursor to next word boundary to the register.
func (e *Editor) yankWord(count int) error {
	row := e.cursor.Row()
	col := e.cursor.Col()
	endRow, endCol := row, col
	for i := 0; i < count; i++ {
		endRow, endCol = e.findWordEnd(endRow, endCol)
	}

	if endRow == row {
		line, _ := e.buffer.GetLine(row)
		if endCol > len(line) {
			endCol = len(line)
		}
		e.register = Register{Content: line[col:endCol], Type: RegisterChar}
	} else {
		lines := e.buffer.GetLines()
		var yanked string
		yanked = lines[row][col:]
		for r := row + 1; r < endRow && r < len(lines); r++ {
			yanked += "\n" + lines[r]
		}
		if endRow < len(lines) {
			yanked += "\n" + lines[endRow][:endCol]
		}
		e.register = Register{Content: yanked, Type: RegisterChar}
	}
	e.setMessage("yanked")
	return nil
}

// copies from cursor to end of line to the register.
func (e *Editor) yankToLineEnd() error {
	line, _ := e.buffer.GetLine(e.cursor.Row())
	col := e.cursor.Col()
	if col < len(line) {
		e.register = Register{Content: line[col:], Type: RegisterChar}
	}
	e.setMessage("yanked")
	return nil
}

// --- Paste operations ---

// pastes after cursor.
func (e *Editor) pasteAfter() {
	e.paste(true)
}

// pastes before cursor.
func (e *Editor) pasteBefore() {
	e.paste(false)
}

// inserts register content
// If afterCursor is true, content goes after the cursor position; otherwise it goes before
func (e *Editor) paste(afterCursor bool) {
	if e.register.Content == "" {
		return
	}

	e.undoMgr.BeginGroup()
	row := e.cursor.Row()
	col := e.cursor.Col()

	if e.register.Type == RegisterLine {
		// Line paste: insert above or below current line
		insertRow := row
		if afterCursor {
			insertRow = row + 1
		}
		pasteLines := strings.Split(e.register.Content, "\n")
		for i, l := range pasteLines {
			e.undoMgr.Record(Action{
				Type:      ActionInsertLine,
				Row:       insertRow + i,
				Text:      l,
				CursorRow: row,
				CursorCol: col,
			})
			e.buffer.InsertLineWithContent(insertRow+i, l)
		}
		e.cursor.MoveTo(insertRow, 0, e.buffer)
	} else {
		// Character paste: insert at or after cursor column
		line, _ := e.buffer.GetLine(row)
		insertCol := col
		if afterCursor {
			insertCol = col + 1
			if insertCol > len(line) {
				insertCol = len(line)
			}
		}
		newLine := line[:insertCol] + e.register.Content + line[insertCol:]
		e.undoMgr.Record(Action{
			Type:      ActionSetLine,
			Row:       row,
			Text:      newLine,
			PrevText:  line,
			CursorRow: row,
			CursorCol: col,
		})
		e.buffer.SetLine(row, newLine)
		e.cursor.MoveTo(row, insertCol+utf8.RuneCountInString(e.register.Content)-1, e.buffer)
	}

	e.undoMgr.EndGroup()
}

// --- Undo/Redo ---

func (e *Editor) undo() {
	group := e.undoMgr.Undo()
	if group == nil {
		e.setMessage("Already at oldest change")
		return
	}

	// Apply inverse in reverse order
	for i := len(group.Actions) - 1; i >= 0; i-- {
		a := group.Actions[i]
		e.applyInverse(a)
	}

	// Restore cursor from the first action's saved position
	if len(group.Actions) > 0 {
		first := group.Actions[0]
		e.cursor.MoveTo(first.CursorRow, first.CursorCol, e.buffer)
	}

	e.setMessage("Undone")
}

func (e *Editor) redo() {
	group := e.undoMgr.Redo()
	if group == nil {
		e.setMessage("Already at newest change")
		return
	}

	// Apply in forward order
	for _, a := range group.Actions {
		e.applyForward(a)
	}

	// Restore cursor to the position saved at the time of the last action
	if len(group.Actions) > 0 {
		last := group.Actions[len(group.Actions)-1]
		e.cursor.MoveTo(last.CursorRow, last.CursorCol, e.buffer)
	}

	e.setMessage("Redone")
}

// reverses a single action directly on the buffer.
func (e *Editor) applyInverse(a Action) {
	switch a.Type {
	case ActionInsertChar:
		e.buffer.DeleteChar(a.Row, a.Col)
	case ActionDeleteChar:
		e.buffer.InsertChar(a.Row, a.Col, rune(a.Text[0]))
	case ActionSplitLine:
		// Reverse split: join the two lines back
		e.buffer.SetLine(a.Row, a.PrevText)
		if a.Row+1 < e.buffer.NumLines() {
			e.buffer.DeleteLine(a.Row + 1)
		}
	case ActionJoinLines:
		// Reverse join: split back into two lines
		e.buffer.SetLine(a.Row, a.Text)
		e.buffer.InsertLineWithContent(a.Row+1, a.PrevText)
	case ActionDeleteLine:
		e.buffer.InsertLineWithContent(a.Row, a.Text)
	case ActionInsertLine:
		e.buffer.DeleteLine(a.Row)
	case ActionSetLine:
		e.buffer.SetLine(a.Row, a.PrevText)
	}
}

// re-applies a single action directly on the buffer.
func (e *Editor) applyForward(a Action) {
	switch a.Type {
	case ActionInsertChar:
		e.buffer.InsertChar(a.Row, a.Col, rune(a.Text[0]))
	case ActionDeleteChar:
		e.buffer.DeleteChar(a.Row, a.Col)
	case ActionSplitLine:
		e.buffer.SplitLine(a.Row, a.Col)
	case ActionJoinLines:
		e.buffer.JoinLines(a.Row)
	case ActionDeleteLine:
		e.buffer.DeleteLine(a.Row)
	case ActionInsertLine:
		e.buffer.InsertLineWithContent(a.Row, a.Text)
	case ActionSetLine:
		e.buffer.SetLine(a.Row, a.Text)
	}
}

// changes the editor mode.
func (e *Editor) setMode(mode Mode) {
	e.mode = mode

	// clear message when changing modes
	if mode != ModeCommand && mode != ModeSearch {
		e.message = ""
	}
}

// sets the message to display in the message bar.
func (e *Editor) setMessage(msg string) {
	e.message = msg
	e.messageTime = time.Now()
}

// updates the scroll offsets to keep the cursor visible.
func (e *Editor) updateScroll() {
	// visible rows (total height - status bar - message bar)
	visibleRows := e.terminal.Height() - 2

	// gutter width to determine available text width
	gutterWidth := ui.GutterWidth(e.buffer.NumLines())
	visibleCols := e.terminal.Width() - gutterWidth

	e.cursor.UpdateScroll(visibleRows, visibleCols)
}

// creates a view struct for rendering.
func (e *Editor) buildView() ui.EditorView {
	// Show command buffer in command mode, search prompt in search mode
	msg := e.message
	if e.mode == ModeCommand {
		msg = e.commandBuf
	} else if e.mode == ModeSearch {
		prefix := "/"
		if e.search.Direction == SearchBackward {
			prefix = "?"
		}
		msg = prefix + e.searchBuf
	}

	view := ui.EditorView{
		Buffer: ui.BufferView{
			Lines:      e.buffer.GetLines(),
			FileName:   e.buffer.FileName(),
			IsModified: e.buffer.IsModified(),
		},
		Cursor: ui.CursorView{
			Row:       e.cursor.Row(),
			Col:       e.cursor.Col(),
			RowOffset: e.cursor.RowOffset(),
			ColOffset: e.cursor.ColOffset(),
		},
		Mode: ui.ModeView{
			Name: e.mode.ShortString(),
		},
		Message:    msg,
		TermWidth:  e.terminal.Width(),
		TermHeight: e.terminal.Height(),
		TotalLines: e.buffer.NumLines(),
	}

	// Search highlighting
	if e.search.Active && len(e.search.Matches) > 0 {
		view.SearchActive = true
		view.SearchMatches = make(map[int][]ui.MatchRange)
		for _, m := range e.search.Matches {
			view.SearchMatches[m.Row] = append(view.SearchMatches[m.Row], ui.MatchRange{
				ColStart: m.ColStart,
				ColEnd:   m.ColEnd,
			})
		}
	}

	// Bracket matching
	match := FindMatchingBracket(e.buffer.GetLines(), e.cursor.Row(), e.cursor.Col())
	if match != nil {
		view.BracketMatch = &ui.BracketMatchView{
			Row: match.Row,
			Col: match.Col,
		}
	}

	return view
}
