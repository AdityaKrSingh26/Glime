package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AdityaKrSingh26/Glime/internal/buffer"
	"github.com/AdityaKrSingh26/Glime/internal/cursor"
	"github.com/AdityaKrSingh26/Glime/internal/terminal"
)

type ExplorerEntry struct {
	Name  string
	IsDir bool
	Size  int64
}

type ExplorerState struct {
	Dir       string
	Entries   []ExplorerEntry
	CursorRow int
	RowOffset int

	savedBuffer *buffer.Buffer
	savedCursor *cursor.Cursor
}

// reads the directory and populates entries, sorted dirs-first then files (case-insensitive).
func (es *ExplorerState) LoadDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	dirEntries, err := os.ReadDir(absDir)
	if err != nil {
		return fmt.Errorf("cannot read directory: %w", err)
	}

	es.Dir = absDir
	es.CursorRow = 0
	es.RowOffset = 0

	// start with parent directory entry
	es.Entries = []ExplorerEntry{{Name: "../", IsDir: true}}

	// split into dirs and files
	var dirs, files []ExplorerEntry
	for _, de := range dirEntries {
		if de.IsDir() {
			dirs = append(dirs, ExplorerEntry{Name: de.Name() + "/", IsDir: true})
		} else {
			size := int64(0)
			if info, err := de.Info(); err == nil {
				size = info.Size()
			}
			files = append(files, ExplorerEntry{Name: de.Name(), Size: size})
		}
	}

	// sort both lists case-insensitively
	byName := func(a, b ExplorerEntry) bool {
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	}
	sort.Slice(dirs, func(i, j int) bool { return byName(dirs[i], dirs[j]) })
	sort.Slice(files, func(i, j int) bool { return byName(files[i], files[j]) })

	es.Entries = append(es.Entries, dirs...)
	es.Entries = append(es.Entries, files...)
	return nil
}

func (es *ExplorerState) EntryAtCursor() *ExplorerEntry {
	if es.CursorRow < 0 || es.CursorRow >= len(es.Entries) {
		return nil
	}
	return &es.Entries[es.CursorRow]
}

func (es *ExplorerState) FullPath() string {
	entry := es.EntryAtCursor()
	if entry == nil {
		return ""
	}
	name := strings.TrimSuffix(entry.Name, "/")
	return filepath.Join(es.Dir, name)
}

func (es *ExplorerState) MoveUp() {
	if es.CursorRow > 0 {
		es.CursorRow--
	}
}

func (es *ExplorerState) MoveDown() {
	if es.CursorRow < len(es.Entries)-1 {
		es.CursorRow++
	}
}

func (es *ExplorerState) UpdateScroll(visibleRows int) {
	if es.CursorRow < es.RowOffset {
		es.RowOffset = es.CursorRow
	}
	if es.CursorRow >= es.RowOffset+visibleRows {
		es.RowOffset = es.CursorRow - visibleRows + 1
	}
}

// handles key input in explore mode.
func (e *Editor) processExploreMode(key *terminal.Key) error {
	switch key.Type {
	case terminal.KeyEscape:
		e.explorerQuit()
	case terminal.KeyEnter:
		e.explorerOpen()
	case terminal.KeyArrowUp:
		e.explorer.MoveUp()
	case terminal.KeyArrowDown:
		e.explorer.MoveDown()
	case terminal.KeyRune:
		e.processExploreRune(key.Rune)
	}
	return nil
}

func (e *Editor) processExploreRune(ch rune) {
	switch ch {
	case 'j':
		e.explorer.MoveDown()
	case 'k':
		e.explorer.MoveUp()
	case 'l':
		e.explorerOpen()
	case 'h', '-':
		e.explorerParent()
	case 'q':
		e.explorerQuit()
	case 'G':
		e.explorer.CursorRow = len(e.explorer.Entries) - 1
	case 'g':
		e.explorer.CursorRow = 0
	case ':':
		e.prevMode = e.mode
		e.setMode(ModeCommand)
		e.commandBuf = ":"
	}
}

// resolveExplorerDir returns the directory to explore:
// uses the given dir, or the current file's directory, or cwd as fallback.
func (e *Editor) resolveExplorerDir(dir string) (string, error) {
	if dir != "" {
		return dir, nil
	}
	if fp := e.buffer.FilePath(); fp != "" {
		return filepath.Dir(fp), nil
	}
	return os.Getwd()
}

// opens the explorer, saving current buffer state.
func (e *Editor) commandExplore(dir string) error {
	dir, err := e.resolveExplorerDir(dir)
	if err != nil {
		return err
	}

	e.explorer.savedBuffer = e.buffer
	e.explorer.savedCursor = e.cursor

	if err := e.explorer.LoadDir(dir); err != nil {
		e.setMessage(fmt.Sprintf("Explore: %v", err))
		return nil
	}

	e.setMode(ModeExplore)
	e.setMessage(fmt.Sprintf("netrw: %s", e.explorer.Dir))
	return nil
}

// opens the selected entry — directory or file.
func (e *Editor) explorerOpen() {
	entry := e.explorer.EntryAtCursor()
	if entry == nil {
		return
	}

	fullPath := e.explorer.FullPath()

	if entry.IsDir {
		if err := e.explorer.LoadDir(fullPath); err != nil {
			e.setMessage(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	// opening a file — warn if saved buffer has unsaved changes
	if e.explorer.savedBuffer != nil && e.explorer.savedBuffer.IsModified() {
		e.setMessage("Unsaved changes! Use :q! or save first")
		return
	}

	// restore cursor so LoadFile works on a fresh state
	e.cursor = cursor.New()
	e.buffer = buffer.New()

	if err := e.LoadFile(fullPath); err != nil {
		e.setMessage(fmt.Sprintf("Error opening file: %v", err))
		return
	}

	e.explorer.savedBuffer = nil
	e.explorer.savedCursor = nil
	e.setMode(ModeNormal)
}

// exits the explorer and restores the previous buffer.
func (e *Editor) explorerQuit() {
	if e.explorer.savedBuffer != nil {
		e.buffer = e.explorer.savedBuffer
		e.cursor = e.explorer.savedCursor
		e.renderer.SetLanguage(e.buffer.FilePath())
		e.explorer.savedBuffer = nil
		e.explorer.savedCursor = nil
	}
	e.setMode(ModeNormal)
	e.setMessage("")
}

// navigates to the parent directory.
func (e *Editor) explorerParent() {
	parent := filepath.Dir(e.explorer.Dir)
	if parent == e.explorer.Dir {
		return // already at root
	}
	if err := e.explorer.LoadDir(parent); err != nil {
		e.setMessage(fmt.Sprintf("Error: %v", err))
	}
}
