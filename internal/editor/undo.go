package editor

// ActionType represent a type of buffer mutation for undo/redo
type ActionType int

const (
	ActionInsertChar ActionType = iota
	ActionDeleteChar
	ActionSplitLine
	ActionJoinLines
	ActionDeleteLine
	ActionInsertLine
	ActionSetLine
)

// record a single buffer mutation
type Action struct {
	Type      ActionType
	Row       int
	Col       int
	Text      string // The text involved (char inserted, line content, etc.)
	PrevText  string // Previous text (for reversal)
	CursorRow int    // Cursor position before the action
	CursorCol int
}

// batches multiple actions into one undoable unit.
type ActionGroup struct {
	Actions []Action
}

// tracks all undo/redo stacks
type UndoManager struct {
	undoStack []ActionGroup
	redoStack []ActionGroup
	current   *ActionGroup
	maxSize   int
}

func NewUndoManager(maxSize int) *UndoManager {
	return &UndoManager{
		undoStack: make([]ActionGroup, 0),
		redoStack: make([]ActionGroup, 0),
		current:   nil,
		maxSize:   maxSize,
	}
}

// starts a new action group for batching.
func (u *UndoManager) BeginGroup() {
	u.current = &ActionGroup{
		Actions: make([]Action, 0),
	}
}

// finalise the current group and push it onto the undo stack
func (u *UndoManager) EndGroup() {
	if u.current == nil || len(u.current.Actions) == 0 {
		u.current = nil
		return
	}

	u.undoStack = append(u.undoStack, *u.current)
	if len(u.undoStack) > u.maxSize {
		u.undoStack = u.undoStack[1:]
	}
	u.current = nil
}

// pops the top group from the undo stack.
// returns nil if nothing to undo.
func (u *UndoManager) Undo() *ActionGroup {
	if len(u.undoStack) == 0 {
		return nil
	}

	// If there's an active group, finalize it first
	if u.current != nil && len(u.current.Actions) > 0 {
		u.EndGroup()
	}

	idx := len(u.undoStack) - 1
	group := u.undoStack[idx]
	u.undoStack = u.undoStack[:idx]
	u.redoStack = append(u.redoStack, group)
	return &group
}

// pops the top group from the redo stack.
// returns nil if nothing to redo.
func (u *UndoManager) Redo() *ActionGroup {
	if len(u.redoStack) == 0 {
		return nil
	}

	idx := len(u.redoStack) - 1
	group := u.redoStack[idx]
	u.redoStack = u.redoStack[:idx]
	u.undoStack = append(u.undoStack, group)
	return &group
}

// check if there are actions to undo.
func (u *UndoManager) CanUndo() bool {
	return len(u.undoStack) > 0 || (u.current != nil && len(u.current.Actions) > 0)
}

// check if there are actions to redo.
func (u *UndoManager) CanRedo() bool {
	return len(u.redoStack) > 0
}

// Store a mutation in history so it can be undone later.
// this adds an action to the current group.
// If no group is active, wraps the action in its own group.
func (u *UndoManager) Record(action Action) {
	// Any new edit destroys the redo history
	// clear redo on new edit
	u.redoStack = u.redoStack[:0]

	if u.current != nil {
		u.current.Actions = append(u.current.Actions, action)
		return
	}

	// auto-wrap in a single-action group
	u.undoStack = append(u.undoStack, ActionGroup{
		Actions: []Action{action},
	})
	if len(u.undoStack) > u.maxSize {
		u.undoStack = u.undoStack[1:]
	}
}
