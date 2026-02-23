package editor

// this package provide the main editor state machine and event loop

import (
	"fmt"
	"os"
	"time"

	"github.com/AdityaKrSingh26/Glime/internal/buffer"
	"github.com/AdityaKrSingh26/Glime/internal/cursor"
	"github.com/AdityaKrSingh26/Glime/internal/file"
	"github.com/AdityaKrSingh26/Glime/internal/input"
	"github.com/AdityaKrSingh26/Glime/internal/terminal"
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
		e.renderer.setLanguage(filePath)
		e.setMessage(fmt.Sprintf("\"%s\" [New File]", filePath))
		return nil
	}

	lines, err := file.Load(filePath)
	if err != nil {
		return fmt.Errorf("Error loading file into editor: %w", err)
	}

	e.buffer = buffer.NewFromLines(lines, filePath)
	e.renderer.setLanguage(filePath)
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

